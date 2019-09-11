variable "vpc-id" {
  description = "Current VPC id"
  type        = string
}

variable "public-subnet-id" {
  description = "ID of the public subnet"
  type        = string
}

variable "frontend-instance-id" {
  description = "Id of the front-end instance"
  type        = string
}

variable "private-subnet-cidr" {
  description = "Cidr of the private subnet where the app server lives"
  type        = string
}

variable "public-subnet-cidr" {
  description = "Cidr of the public subnet where LB lives"
  type        = string
}