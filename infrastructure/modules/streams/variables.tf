variable "stream_name" {
  description = "Name of the Kinesis stream"
  type        = string
}

variable "shard_count" {
  description = "The number of shards for the stream"
  type        = number
  default     = 2
}
