variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "db_username" {
  type = string
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "bucket_name" {
  description = "The S3 bucket name for telemetry archive"
  type        = string
}

variable "terraform_state_bucket" {
  description = "S3 bucket for Terraform remote state storage"
  type        = string
}

variable "terraform_lock_table" {
  description = "DynamoDB table for Terraform state locking"
  type        = string
}
