terraform {
  backend "s3" {
    bucket         = "ee-tf-remote-state-kkittel"
    key            = "staging/terraform.tfstate"
    region         = "eu-west-2"
    dynamodb_table = "ee-tf-locks-kkittel"
    encrypt        = true
  }
}

provider "aws" {
  region = var.region
}

locals {
  main-vpc-cidr  = "10.10.0.0/16"
  public-subnet-cidr = "10.10.1.0/24"
  private-subnet-cidr = "10.10.2.0/24"
}

resource "aws_key_pair" "kasia" {
  key_name   = "kasia-key"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDq3K5gel1+OYYF3pjXx2Wn+eV8s1VLMQd6b0iEmurR5kFubUYQyZ0iZVlM1717uxLS5wDxLJs/nySaI9KB5iEsqxlx/NTe5uP6i7Rd2z5r6faGn/S2aY7xZZcGaGZsuduM6UG9xvlSAJ9IPCRS/vUw3d8xDDxflwYqGtmfVWXgOm4I871vN7NrpGH7GcCFuyDx/2PV/1Oh/e5OAVWMvXrz/X2GJKUEgxa7VwWsxulapcfr8QiGmSi6raHiHWeeujGZi27TtFcv+dhUrbfz+B/fgAK7wPr0bDzG2FxoFArs1Q8Vgb+v6w2U46ftM+S94It2o34oV1e6lsFIcqIdcaDtvGasVz0aoh2OWWUoiCorgCVIcMcEZn+/XJKbP0e9g4REgcu1gu67/5Mxc5VULofNKkwjKZwMk4uhDeem7LIFERVKB7xXbQF57e/qO2ZkhYVR5J+hM5hIB49MiVKzn5MFIRzYGsqJogT3zxct6lVBc/e0OHbcYAwoN/rnadgTfpS4gq2q1Iql3kLIbYFoy9oaxPxT0R3H30CbhlQitgR6WarkmQrTL8rpVtcjALuPToiEL3wcu/3Wae7RNviHo9huFmmZa0IMeS0CCmvScIJGTM9XpflEaFUFeH2LcI3Kq1vPDgq5slMb94CM/V0zHyKl6Oa0oiNwwLBfIPRU4ZSrQQ== kasia@Katarzynas-MBP"
}

module "aws-ec2-bastion" {
  source = "../modules/aws-ec2-generic"
  aws-ec2-name  = "bastion"
  vpc-sg-ids = [module.aws-sg.public-ssh-sg-id]
  subnet-id = module.aws-network.public-subnet-id
  key-name = aws_key_pair.kasia.key_name
}

module "aws-ec2-frontend" {
  source = "../modules/aws-ec2-generic"
  aws-ec2-name  = "frontend"
  vpc-sg-ids = [module.aws-sg.private-ssh-sg-id, module.aws-sg.frontend-http-sg-id]
  subnet-id = module.aws-network.public-subnet-id
  key-name = aws_key_pair.kasia.key_name
}

module "aws-ec2-backend" {
  source = "../modules/aws-ec2-generic"
  aws-ec2-name  = "backend"
  vpc-sg-ids = [module.aws-sg.private-ssh-sg-id, module.aws-sg.backend-http-sg-id]
  subnet-id = module.aws-network.private-subnet-id
  key-name = aws_key_pair.kasia.key_name 
}

module "aws-network" {
  source = "../modules/aws-network"
  main-vpc-cidr  = local.main-vpc-cidr 
  public-subnet-cidr = local.public-subnet-cidr
  private-subnet-cidr = local.private-subnet-cidr
}

module "aws-sg" {
  source = "../modules/aws-sg"
  vpc-id = module.aws-network.main-vpc-id
  public-subnet-cidr = local.public-subnet-cidr
  private-subnet-cidr = local.private-subnet-cidr
}

module "aws-nat" {
  source = "../modules/aws-nat"
  vpc-id = module.aws-network.main-vpc-id
  private-subnet-id = module.aws-network.private-subnet-id
  public-subnet-id = module.aws-network.public-subnet-id
  public-subnet-cidr = local.public-subnet-cidr
  private-subnet-cidr = local.private-subnet-cidr
}

output "bastion-ip" {
  value = module.aws-ec2-bastion.instance-public-ip
}

output "frontend-public-ip" {
  value = module.aws-ec2-frontend.instance-public-ip
}

output "frontend-private-ip" {
  value = module.aws-ec2-frontend.instance-private-ip
}

output "backend-private-ip" {
  value = module.aws-ec2-backend.instance-private-ip
}

output "nat-ip" {
  value = module.aws-nat.nat-ip
}