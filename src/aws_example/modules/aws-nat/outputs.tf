output "nat-ip" {
  value = aws_instance.nat.public_ip
}

output "nat-instance-id" {
  value = aws_instance.nat.id
}