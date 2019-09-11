terraform {
    required_version = ">= 0.12"
}

resource "aws_instance" "aws-ec2" {
  ami           =  data.aws_ami.ubuntu.id
  instance_type = "t2.micro"

  vpc_security_group_ids = var.vpc-sg-ids
  subnet_id = var.subnet-id
  key_name = var.key-name

  tags = {
    Project = "terraform-example-kasia"
    Name = var.aws-ec2-name
  }
  availability_zone = var.availability-zone
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "image-type"
    values = ["machine"]
  }

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"]
  }
}