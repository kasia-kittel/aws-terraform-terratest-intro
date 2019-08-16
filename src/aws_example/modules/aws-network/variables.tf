variable "main-vpc-cidr" {
  description = "The CIDR of the main VPC"
  type        = string
}

variable "main-vpc-name" {
  description = "Name tag of the main VPC"
  type = string
  default = "Main VPC"
}

variable "default-igw-name" {
  description = "Name tag of the main IGW"
  type = string
  default = "Main VPC - default IGW"
}

variable "public-subnet-cidr" {
  description = "The CIDR of public subnet"
  type        = string
}

variable "public-subnet-name" {
  description = "Name tag of the public subnet"
  type = string
  default = "Public Subnet"
}

variable "private-subnet-cidr" {
  description = "The CIDR of the private subnet"
  type        = string
}

variable "private-subnet-name" {
  description = "Name tag of the private subnet"
  type = string
  default = "Private Subnet"
}
