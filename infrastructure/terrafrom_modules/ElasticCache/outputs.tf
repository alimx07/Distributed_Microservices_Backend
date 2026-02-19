output "cache_connections" {
  value = {
    for name, svc in local.services_map : name => merge(
      {
        secret_arn = aws_secretsmanager_secret.cache_credentials[name].arn
      },
      svc.cluster_mode ? {
        cluster_endpoint = aws_elasticache_replication_group.cluster[name].configuration_endpoint_address
      } : {},
      svc.standalone_mode ? {
        standalone_endpoint = aws_elasticache_replication_group.standalone[name].primary_endpoint_address
      } : {}
    )
  }
  sensitive = true
}

output "security_group_ids" {
  value = module.sg.security_group_ids
}
