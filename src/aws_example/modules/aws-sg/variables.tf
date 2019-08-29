variable "vpc-id" {
  description = "The Id of the VPC"
  type        = string
}

variable "public-subnet-cidr" {
  description = "The CIDR of the public subnet"
  type        = string
}

variable "private-subnet-cidr" {
  description = "The CIDR of the private subnet"
  type        = string
}