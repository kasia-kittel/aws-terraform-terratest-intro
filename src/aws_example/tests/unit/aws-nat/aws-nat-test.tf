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

resource "aws_vpc" "nat-test" {
  cidr_block =  "10.20.0.0/16"

  tags = {
    Project = "terraform-example-kasia"
    Name = "nat-test"
  }
}

resource "aws_subnet" "nat-test-private" {
  vpc_id      = aws_vpc.nat-test.id
  cidr_block  = var.private-subnet-cidr
}

resource "aws_subnet" "nat-test-public" {
  vpc_id      = aws_vpc.nat-test.id
  cidr_block  = var.public-subnet-cidr
}

module "aws-nat" {
  source = "../../../modules/aws-nat"
  vpc-id = aws_vpc.nat-test.id
  private-subnet-id = aws_subnet.nat-test-private.id
  public-subnet-id = aws_subnet.nat-test-public.id
  public-subnet-cidr = var.public-subnet-cidr
  private-subnet-cidr = var.private-subnet-cidr
}

variable "region" {
  type = string
}

variable "private-subnet-cidr" {
  type = string
}

variable "public-subnet-cidr" {
  type = string
}

output "nat-ip" {
  value = module.aws-nat.nat-ip
}

output "nat-instance-id" {
  value = module.aws-nat.nat-instance-id
}

output "test-vpc-id" {
  value = aws_vpc.nat-test.id
}

output "private-subnet-id" {
  value = aws_subnet.nat-test-private.id
}

