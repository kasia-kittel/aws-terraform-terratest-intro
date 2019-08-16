terraform {
  backend "s3" {
    bucket         = "ee-tf-remote-state-kkittel"
    key            = "test/aws-ec2-generic/terraform.tfstate"
    region         = "eu-west-2"
    dynamodb_table = "ee-tf-locks-kkittel"
    encrypt        = true
  }
}

provider "aws" {
  region = var.region
}

module "aws-network-test" {
  source = "../../../modules/aws-network"
  main-vpc-cidr  = var.main-vpc-cidr
  main-vpc-name = var.main-vpc-name
  default-igw-name = "default-igw-test "
  public-subnet-cidr = var.public-subnet-cidr
  public-subnet-name = var.public-subnet-name
  private-subnet-cidr = var.private-subnet-cidr
  private-subnet-name = var.private-subnet-name
}

output "main-vpc-id" {
  value = module.aws-network-test.main-vpc-id
}

output "public-subnet-id" {
  value = module.aws-network-test.public-subnet-id
}

output "private-subnet-id" {
  value = module.aws-network-test.private-subnet-id
}

output "default-igw-id" {
  value = module.aws-network-test.default-igw-id
}

output "public-ssh-sg-id" {
  value = module.aws-network-test.public-ssh-sg-id
}

output "private-ssh-sg-id" {
  value = module.aws-network-test.private-ssh-sg-id
}

variable "region" {
  description = "The AWS region"
  type        = string
  default     = "eu-west-2"
}

variable "main-vpc-cidr" {
  description = "The CIDR of the main VPC"
  type        = string
  default     = "10.10.0.0/16"
}

variable "main-vpc-name" {
  description = "The name of the main VPC"
  type        = string
  default     = "main-vpc-test"
}

variable "public-subnet-cidr" {
  description = "The CIDR of public subnet"
  type        = string
  default     = "10.10.1.0/24"
}

variable "public-subnet-name" {
  description = "Name tag of the public subnet"
  type        = string
  default     = "public-subnet-test"
}

variable "private-subnet-cidr" {
  description = "The CIDR of public subnet"
  type        = string
  default     = "10.10.2.0/24"
}

variable "private-subnet-name" {
  description = "Name tag of the public subnet"
  type        = string
  default     = "private-subnet-test"
}
