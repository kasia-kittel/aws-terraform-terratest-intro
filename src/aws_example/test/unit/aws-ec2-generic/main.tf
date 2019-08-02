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
  region = "eu-west-2"
}

module "aws-ec2-generic-test" {
  source = "../../../modules/aws-ec2-generic"
  aws-ec2-name  = var.aws-ec2-name
}

output "aws-ec2-id" {
  value = module.aws-ec2-generic-test.aws-ec2-id
}

variable "aws-ec2-name" {
  description = "The name of the EC2 instance"
  type        = string
  default = "generic - test"
}
