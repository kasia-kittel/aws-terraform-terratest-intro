output "frontend-http-sg-id" {
  value = aws_security_group.frontend-http.id
  description = "Public Http access to front-end service"
}

output "backend-http-sg-id" {
  value = aws_security_group.backend-http.id
  description = "Http access only from public subnetwork"
}

output "public-ssh-sg-id" {
  value       = aws_security_group.public-ssh.id
  description = "Ssh access security group"
}

output "private-ssh-sg-id" {
  value       = aws_security_group.private-ssh.id
  description = "Ssh from vpc network - security group"
}