provider "aws" {
  region = var.aws_region
}

# Kinesis Data Stream for telemetry ingestion.
resource "aws_kinesis_stream" "telemetry_stream" {
  name        = "iot-telemetry-stream"
  shard_count = 2
}

# RDS Instance with TimescaleDB (PostgreSQL with Timescale extension).
resource "aws_db_instance" "timescaledb" {
  allocated_storage    = 20
  engine               = "postgres"
  engine_version       = "13.3"
  instance_class       = "db.t3.medium"
  name                 = "telemetrydb"
  username             = var.db_username
  password             = var.db_password
  skip_final_snapshot  = true

  # Additional parameter groups needed to enable TimescaleDB extension.
}

# S3 Bucket for archiving raw telemetry logs.
resource "aws_s3_bucket" "telemetry_archive" {
  bucket = "iot-telemetry-archive-${random_id.bucket_id.hex}"
}

resource "random_id" "bucket_id" {
  byte_length = 4
}

# AWS Lambda for event-driven alerting.
resource "aws_lambda_function" "alerting_function" {
  function_name = "iot_alerting"
  handler       = "alerting_function"
  runtime       = "go1.x"
  role          = aws_iam_role.lambda_exec.arn
  filename      = "lambda/alerting.zip"  # This is your built Lambda package.
}

resource "aws_iam_role" "lambda_exec" {
  name = "lambda_exec_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

# Outputs for integration.
output "kinesis_stream_name" {
  value = aws_kinesis_stream.telemetry_stream.name
}
output "rds_endpoint" {
  value = aws_db_instance.timescaledb.endpoint
}
