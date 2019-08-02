output "aws-ec2-id" {
  value = aws_instance.aws-ec2.id
  description = "The EC2 instance ID"
}