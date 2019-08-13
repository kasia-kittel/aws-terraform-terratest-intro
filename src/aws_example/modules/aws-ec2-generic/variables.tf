variable "aws-ec2-name" {
  description = "The name of the EC2 instance"
  type        = string
  default     = "generic"
}

variable "vpc_sg_id" {
  description = "Security group id. Use default security group if not set"
  type        = list(string)
  default     = null
}

variable "subnet_id" {
  description = "Subnet id. Use default subnet if not set"
  type        = string
  default     = null
}

variable "key_name" {
  description = "Key name for SSH session."
  type        = string
  default     = null
}