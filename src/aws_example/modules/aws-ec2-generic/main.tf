terraform {
    required_version = ">= 0.12"
}

resource "aws_instance" "aws-ec2" {
  ami           = "ami-025f234ba1c577d83"
  instance_type = "t2.micro"

  vpc_security_group_ids = var.vpc_sg_id

  tags = {
    Project = "terraform-example-kasia"
    Name = var.aws-ec2-name
  }
}