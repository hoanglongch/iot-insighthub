provider "aws" {
  region = var.aws_region
}

module "streams" {
  source      = "./modules/streams"
  stream_name = "iot-telemetry-stream"
  shard_count = 2
}

module "database" {
  source           = "./modules/database"
  db_name          = "telemetrydb"
  db_username      = var.db_username
  db_password      = var.db_password
  engine_version   = "13.3"
  instance_class   = "db.t3.medium"
  allocated_storage = 20
}

module "storage" {
  source      = "./modules/storage"
  bucket_name = var.bucket_name
}

# AWS Lambda and IAM resources can remain here or be modularized in a similar fashion.
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
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })
}
