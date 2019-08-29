terraform {
  required_version = ">= 0.12"
}

resource "aws_vpc" "main" {
  cidr_block =  var.main-vpc-cidr

  tags = {
    Project = "terraform-example-kasia"
    Name = var.main-vpc-name
  }
}

resource "aws_internet_gateway" "default" {
  vpc_id = aws_vpc.main.id

  tags = {
    Project = "terraform-example-kasia"
    Name = var.default-igw-name
  }
}

resource "aws_subnet" "public" {
    vpc_id = aws_vpc.main.id

    cidr_block = var.public-subnet-cidr
    
    map_public_ip_on_launch = true
    
    tags = {
        Project = "terraform-example-kasia"
        Name = var.public-subnet-name
    }
}

resource "aws_subnet" "private" {
    vpc_id = aws_vpc.main.id

    cidr_block = var.private-subnet-cidr
    
    map_public_ip_on_launch = false
    
    tags = {
        Project = "terraform-example-kasia"
        Name = var.private-subnet-name
    }
}

resource "aws_route_table" "public" {
    vpc_id = aws_vpc.main.id

    route {
        cidr_block = "0.0.0.0/0" #destination
        gateway_id = aws_internet_gateway.default.id
    }

    tags = {
        Project = "terraform-example-kasia"
        Name = "Public Subnet Route"
    }
}

resource "aws_route_table_association" "public" {
    subnet_id = aws_subnet.public.id
    route_table_id = aws_route_table.public.id
}
