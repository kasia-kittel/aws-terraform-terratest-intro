terraform {
  required_version = ">= 0.12"
}

locals {
  s3-backet-prefix    = "lb-logs"
  s3-backet-name      = "ee-tf-lb-logs-kkittel"
}

resource "aws_lb" "load-balancer" {
  name               = "terraform-example-load-balancer"
  internal           = false
  load_balancer_type = "network"
  subnets            = [var.public-subnet-id]

  // TODO there is a bug in terraform - it can't remove the lb and the whole destroy fails
  enable_deletion_protection = false

  access_logs {
    bucket  = aws_s3_bucket.lb-logs.id
    prefix  = local.s3-backet-prefix
    enabled = true
  }

  tags = {
    Project = "terraform-example-kasia"
  }
}

data "aws_elb_service_account" "main" {}

// https://docs.aws.amazon.com/elasticloadbalancing/latest/network/load-balancer-access-logs.html
data "aws_iam_policy_document" "s3_lb_write" {
    policy_id = "s3_lb_write"

    statement {
        actions = ["s3:PutObject"]
        resources = ["arn:aws:s3:::ee-tf-lb-logs-kkittel/${local.s3-backet-prefix}/AWSLogs*"]

        principals {
          type = "Service"
          identifiers = [ "delivery.logs.amazonaws.com" ]
        }
    }
}

resource "aws_s3_bucket" "lb-logs" {
  bucket = local.s3-backet-name
  versioning {
    enabled = true
  }

  policy = data.aws_iam_policy_document.s3_lb_write.json

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}

resource "aws_lb_listener" "webapps-listener" {
  load_balancer_arn = aws_lb.load-balancer.arn
  port              = "80"
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.webapps-target-group.arn
  }
}

resource "aws_lb_target_group" "webapps-target-group" {
  name     = "terraform-example-target-group"
  port     = 80
  protocol = "TCP"
  vpc_id   = var.vpc-id

  health_check {
    healthy_threshold   = 10
    unhealthy_threshold = 10
    interval            = 30
    protocol            = "TCP"
  }

  tags = {
      Project = "terraform-example-kasia"
      Name = "webapps-target-group"
  }
}

resource "aws_lb_target_group_attachment" "frontend" {
  target_group_arn = aws_lb_target_group.webapps-target-group.arn
  target_id        = var.frontend-instance-id
  port             = 80
}

// the group to be used by trarget instances
// as recommended: https://docs.aws.amazon.com/elasticloadbalancing/latest/network/target-group-register-targets.html
resource "aws_security_group" "lb-tcp" {
  name      = "terraform-example-lb-tcp"
  description = "Allow TCP inbound and outbound Internet traffic via Load Balancer."
  
  vpc_id = var.vpc-id

  ingress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
      from_port = 80
      to_port = 80
      protocol = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
      Project = "terraform-example-kasia"
      Name = "lb-tcp-security-group"
  }
}

