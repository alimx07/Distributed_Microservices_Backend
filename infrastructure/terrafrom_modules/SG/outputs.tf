output "security_group_ids" {
  description = "List of security group IDs"
  value       = aws_security_group.this[*].id
}

# TODO : Make it output map of sg
