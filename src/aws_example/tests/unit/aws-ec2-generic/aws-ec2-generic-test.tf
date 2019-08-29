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

module "aws-ec2-generic-test" {
  source = "../../../modules/aws-ec2-generic"
  aws-ec2-name  = var.aws-ec2-name
}

output "aws-ec2-id" {
  value = module.aws-ec2-generic-test.aws-ec2-id
}

output "instance-public-ip" {
  value = module.aws-ec2-generic-test.instance-public-ip
}

variable "aws-ec2-name" {
  type      = string
  default   = "generic - test"
}

variable "region" {
  type = string
}