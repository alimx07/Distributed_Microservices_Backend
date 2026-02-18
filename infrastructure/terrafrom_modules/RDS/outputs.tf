output "db_connections" {
  value = {
    for name, svc in local.services_map : name => {
      primary_endpoint = aws_db_instance.primary[name].endpoint
      primary_host     = aws_db_instance.primary[name].address
      replica_endpoint = aws_db_instance.replica[name].endpoint
      replica_host     = aws_db_instance.replica[name].address
      port             = 5432
      db_name          = svc.db_name
      username         = name
      secret_arn       = aws_secretsmanager_secret.db_credentials[name].arn
    }
  }
}

