terraform {
  backend "s3" {}
}



# module "sg" {
#     source = "../SG"
    
# }



# Rand password
resource "random_password" "db_password" {
  for_each = local.services_map

  length           = 24
  special          = true
  override_special = "!#$%&*()-_=+[]{}|:,.<>?"
}


# Secure yourself dude

resource "aws_secretsmanager_secret" "db_credentials" {
  for_each = local.services_map

  name                    = "${local.prefix}/${each.key}/db-credentials"
  recovery_window_in_days = 0

  tags = local.default_tags
}

resource "aws_secretsmanager_secret_version" "db_credentials" {
  for_each = local.services_map

  secret_id = aws_secretsmanager_secret.db_credentials[each.key].id
  secret_string = jsonencode({
    username         = each.key
    password         = random_password.db_password[each.key].result
    db_name          = each.value.db_name
    port             = 5432
    primary_endpoint = aws_db_instance.primary[each.key].endpoint
    primary_host     = aws_db_instance.primary[each.key].address
    replica_endpoint = aws_db_instance.replica[each.key].endpoint
    replica_host     = aws_db_instance.replica[each.key].address
  })
}


## Security Group
module "sg" {
    source = "../SG"
    vpc_id = var.vpc_id
        security_groups = [
        {
            name = "db-sg"
            description = "Security Group for DB"
            ingress_rules = [ {
              cidr_block = var.allowed_cidr_block
                from_port                    = 5432
                to_port                      = 5432
                ip_protocol                  = "tcp"
                description                  = "INTO PostgreSQL SG"
            }],
            egress_rules = [
                {
          cidr_block  = "0.0.0.0/0"
          ip_protocol = "-1"
          description = "OUT PostgreSQL SG"
                }
            ]
        }
    ]
}


# Parameter group for logical replication 
resource "aws_db_parameter_group" "logical_replication" {
  for_each = local.logical_replication_services

  name   = "${local.prefix}-${each.key}-logical-rep"
  family = "postgres${var.engine_version}"

  parameter {
    name         = "rds.logical_replication"
    value        = "1"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "max_replication_slots"
    value = "5"
    apply_method = "pending-reboot"
  }

  parameter {
    name  = "max_wal_senders"
    value = "5"
    apply_method = "pending-reboot"
  }

  tags = merge(local.default_tags, {
    Service = each.key
  })
}

resource "aws_db_subnet_group" "db" {
  subnet_ids = var.subnet_ids
  name = "db_group"
}


# RDS
resource "aws_db_instance" "primary" {
  for_each = local.services_map

  identifier     = "${local.prefix}-${each.key}-primary"
  engine         = "postgres"
  engine_version = var.engine_version
  instance_class = var.instance_class

  allocated_storage = var.allocated_storage
  storage_encrypted = true

  db_name  = each.value.db_name
  username = each.key
  password = random_password.db_password[each.key].result
  port     = 5432

  parameter_group_name = each.value.enable_logical_replication ? aws_db_parameter_group.logical_replication[each.key].name : null


  db_subnet_group_name   = aws_db_subnet_group.db.id
  vpc_security_group_ids = module.sg.security_group_ids

  multi_az            = false
  publicly_accessible = false
  skip_final_snapshot = true

  backup_retention_period = 7

  tags = merge(local.default_tags, {
    Service = each.key
    Role    = "primary"
  })
}


resource "aws_db_instance" "replica" {
  for_each = local.services_map

  identifier     = "${local.prefix}-${each.key}-replica"
  engine         = "postgres"
  engine_version = var.engine_version
  instance_class = var.instance_class

  replicate_source_db = aws_db_instance.primary[each.key].identifier
  storage_encrypted   = true

  vpc_security_group_ids = module.sg.security_group_ids

  multi_az            = false
  publicly_accessible = false
  skip_final_snapshot = true

  tags = merge(local.default_tags, {
    Service = each.key
    Role    = "replica"
  })
}