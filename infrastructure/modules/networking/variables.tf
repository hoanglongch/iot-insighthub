variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
}

variable "vpc_name" {
  description = "Name for the VPC"
  type        = string
}

variable "public_subnet_cidrs" {
  description = "List of CIDRs for public subnets"
  type        = list(string)
}

variable "availability_zones" {
  description = "List of availability zones to use"
  type        = list(string)
}

variable "allowed_db_access_cidrs" {
  description = "CIDRs allowed to access the RDS instance"
  type        = list(string)
  default     = ["10.0.0.0/16"]
}

variable "public_route_table_ids" {
  description = "Route table IDs for VPC endpoints"
  type        = list(string)
}