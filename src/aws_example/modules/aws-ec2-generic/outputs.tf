output "aws-ec2-id" {
  value = aws_instance.aws-ec2.id
  description = "The EC2 instance ID"
}

output "instance_public_ip" {
  value = aws_instance.aws-ec2.public_ip
}

output "instance_private_ip" {
  value = aws_instance.aws-ec2.private_ip
}
