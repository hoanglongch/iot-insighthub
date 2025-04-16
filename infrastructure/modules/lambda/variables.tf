variable "lambda_function_name" {
  description = "Name of the Lambda function"
  type        = string
}

variable "lambda_handler" {
  description = "The handler for the Lambda function"
  type        = string
}

variable "lambda_runtime" {
  description = "The runtime for the Lambda function"
  type        = string
  default     = "go1.x"
}

variable "lambda_role_arn" {
  description = "The ARN of the Lambda IAM role"
  type        = string
}

variable "lambda_filename" {
  description = "The filename of the built Lambda package"
  type        = string
}
