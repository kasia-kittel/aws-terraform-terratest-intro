variable "aws-ec2-name" {
  description = "The name of the EC2 instance"
  type        = string
  default     = "generic"
}

variable "vpc_sg_id" {
  description = "Security group id. Use default security group if not available"
  type        = list(string)
  default     = null
}