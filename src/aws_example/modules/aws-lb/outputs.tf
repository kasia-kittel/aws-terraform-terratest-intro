output "lb-tcp-sg-id" {
  value = aws_security_group.lb-tcp.id
  description = "TCP access via load balancer"
}

output "lb-dns-name" {
  value = aws_lb.load-balancer.dns_name
  description = "TCP access via load balancer"
}
