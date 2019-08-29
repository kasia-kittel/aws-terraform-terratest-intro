variable "private-subnet-cidr" {
  description = "Cidr of the private subnet behind the NAT"
  type        = string
}

variable "private-subnet-id" {
  description = "ID of the private subnet behind the NAT"
  type        = string
}

variable "public-subnet-cidr" {
  description = "Cidr of the public subnet where the NAT instance lives"
  type        = string
}

variable "public-subnet-id" {
  description = "ID of the public subnet where the NAT instance lives"
  type        = string
}

variable "vpc-id" {
  description = "Current VPC id"
  type        = string
}
