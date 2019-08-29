terraform {
  required_version = ">= 0.12"
}

resource "aws_security_group" "frontend-http" {
  name      = "terraform-example-frotnend-http"
  description = "Allow HTTP inbound and outbound Internet traffic."
  
  vpc_id = var.vpc-id

  ingress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  // the outgoing traffic is OK, but could be also
  // limited to some subnetwork if there is no updates 
  // incomming from public networks which is also advisable
  // cidr_blocks = [var.private-subnet-cidr]
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

  egress {
    from_port   = 80
    to_port   = 80
    protocol  = "tcp"
  }
  
  tags = {
      Project = "terraform-example-kasia"
      Name = "frontend-http-security-group"
  }
}

resource "aws_security_group" "backend-http" {
  name      = "terraform-example-backend-http"
  description = "Allow HTTP inbound from the public subnet."
  
  vpc_id = var.vpc-id

  ingress {
      from_port = 80
      to_port = 80
      protocol = "tcp"
      cidr_blocks = [var.public-subnet-cidr] //TODO is it possible to narrow it down to single IP of front-end?
  }

  // these connections should be done via NAT
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
      Project = "terraform-example-kasia"
      Name = "frontend-http-security-group"
  }
}


