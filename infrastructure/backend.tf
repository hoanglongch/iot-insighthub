terraform {
  backend "s3" {
    bucket         = var.terraform_state_bucket
    key            = "iot-insighthub/terraform.tfstate"
    region         = var.aws_region
    dynamodb_table = var.terraform_lock_table
    encrypt        = true
  }
}
