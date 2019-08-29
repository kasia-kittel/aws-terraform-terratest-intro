terraform {
  required_version = ">= 0.12"
}

//TODO add ssh access from bastion

// https://docs.aws.amazon.com/vpc/latest/userguide/VPC_NAT_Instance.html#NATSG
resource "aws_security_group" "nat" {

  name = "terraform-example-nat"
  description = "Allow traffic to pass from the private subnet to the Internet"

  vpc_id = var.vpc-id

  ingress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = [var.private-subnet-cidr]
  }

  ingress {
    from_port = 443
    to_port = 443
    protocol = "tcp"
    cidr_blocks = [var.private-subnet-cidr]
  }

  egress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 443
    to_port = 443
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "NATSG"
  }
}

resource "aws_instance" "nat" {
  ami   =  data.aws_ami.nat-ami.id
  instance_type = "t2.micro"

  vpc_security_group_ids = [aws_security_group.nat.id]
  subnet_id = var.public-subnet-id
  
  // https://docs.aws.amazon.com/vpc/latest/userguide/VPC_NAT_Instance.html#EIP_Disable_SrcDestCheck
  source_dest_check = false

  tags = {
    Project = "terraform-example-kasia"
    Name = "NAT"
  }
}

data "aws_ami" "nat-ami" {
  most_recent = true
  owners  = ["137112412989"] # Amazon

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
    values = ["amzn-ami-vpc-nat-*"]
  }
}

resource "aws_route_table" "nat" {
  vpc_id = var.vpc-id

  route {
    cidr_block = "0.0.0.0/0"
    instance_id = aws_instance.nat.id
  }

  tags = {
    Project = "terraform-example-kasia"
    Name = "NAT Route"
  }
}

resource "aws_route_table_association" "private-nat" {
  subnet_id       = var.private-subnet-id
  route_table_id  = aws_route_table.nat.id
}
