output "kinesis_stream_name" {
  value = module.streams.stream_name
}

output "rds_endpoint" {
  value = module.database.db_endpoint
}

output "s3_bucket" {
  value = module.storage.bucket_name
}
