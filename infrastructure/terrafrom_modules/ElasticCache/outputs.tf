output "cluster_endpoint" {
  description = "Redis cluster configuration endpoint"
  value       = aws_elasticache_replication_group.cluster.configuration_endpoint_address
}

output "cluster_port" {
  value = 6379
}

output "standalone_endpoint" {
  description = "Redis standalone primary endpoint"
  value       = aws_elasticache_replication_group.standalone.primary_endpoint_address
}

output "standalone_port" {
  value = 6379
}

output "security_group_ids" {
  value = module.sg.security_group_ids
}
