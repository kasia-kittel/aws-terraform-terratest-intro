terraform {
    required_version = ">= 0.12"
}

resource "aws_instance" "aws-ec2" {
  ami           = "ami-025f234ba1c577d83"
  instance_type = "t2.micro"

  tags = {
    Project = "terraform-example-kasia"
    Name = var.aws-ec2-name
  }
}