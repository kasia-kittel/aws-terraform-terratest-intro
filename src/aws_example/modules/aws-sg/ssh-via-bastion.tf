terraform {
  required_version = ">= 0.12"
}

resource "aws_security_group" "public-ssh" {
  name      = "terraform-example-public-ssh"
  description = "Allow SSH inbound traffic from the Internet."
  
  vpc_id = var.vpc-id

  ingress {
      from_port = 22
      to_port = 22
      protocol = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port   = 0
    protocol  = "-1"
    cidr_blocks = [var.private-subnet-cidr, var.public-subnet-cidr]
  }
  
  tags = {
      Project = "terraform-example-kasia"
      Name = "public-ssh-security-group"
  }
}

resource "aws_security_group" "private-ssh" {
  name      = "terraform-example-allow_private-ssh"
  description = "Allow SSH inbound from the public subnet."
  
  vpc_id = var.vpc-id

  ingress {
      from_port = 22
      to_port = 22
      protocol = "tcp"
      cidr_blocks = [var.public-subnet-cidr] //TODO is it possible to narrow it down to single IP of bastion?
  }
  
  tags = {
      Project = "terraform-example-kasia"
      Name = "private-ssh-security-group"
  }
}
