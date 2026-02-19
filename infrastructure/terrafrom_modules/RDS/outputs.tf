output "db_connections" {
  value = {
    for name, svc in local.services_map : name => {
      db_name          = svc.db_name
      secret_arn       = aws_secretsmanager_secret.db_credentials[name].arn
    }
  }
  sensitive = true
}


