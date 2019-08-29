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


resource "aws_vpc" "sg-test" {
  cidr_block =  var.vpc-cidr

  tags = {
    Project = "terraform-example-kasia"
    Name = "sg-test"
  }
}

resource "aws_subnet" "sg-test-private" {
  vpc_id      = aws_vpc.sg-test.id
  cidr_block  = var.private-subnet-cidr
}

resource "aws_subnet" "sg-test-public" {
  vpc_id      = aws_vpc.sg-test.id
  cidr_block  = var.public-subnet-cidr
}

module "aws-sg" {
  source = "../../../modules/aws-sg"
  vpc-id = aws_vpc.sg-test.id
  public-subnet-cidr = var.public-subnet-cidr
  private-subnet-cidr = var.private-subnet-cidr
}


variable "vpc-cidr" {
  type = string
}

variable "private-subnet-cidr" {
  type = string
}

variable "public-subnet-cidr" {
  type = string
}

variable "region" {
  type = string
}

output "public-ssh-sg-id" {
  value       = module.aws-sg.public-ssh-sg-id
  description = "Ssh access security group"
}

output "private-ssh-sg-id" {
  value       = module.aws-sg.private-ssh-sg-id
  description = "Ssh from vpc network - security group"
}

output "vpc-id" {
  value = aws_vpc.sg-test.id
}
