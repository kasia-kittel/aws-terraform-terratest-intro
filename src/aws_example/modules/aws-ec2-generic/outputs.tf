output "aws-ec2-id" {
  value = aws_instance.aws-ec2.id
  description = "The EC2 instance ID"
}

output "instance-public-ip" {
  value = aws_instance.aws-ec2.public_ip
}

output "instance-private-ip" {
  value = aws_instance.aws-ec2.private_ip
}
