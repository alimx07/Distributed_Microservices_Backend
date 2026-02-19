terraform {
  backend "s3" {}
}

data "aws_caller_identity" "caller_id" {}


module "sg" {
  source = "../SG"
  vpc_id = var.vpc_id
  security_groups = [
    {
      name        = "${local.prefix}-redis-sg"
      description = "ElastiCache Redis access"
      ingress_rules = [
        {
          cidr_block  = var.private_cidr
          from_port   = 6379
          to_port     = 6379
          ip_protocol = "tcp"
          description = "Redis from private subnets"
        }
      ]
      egress_rules = [
        {
          cidr_block  = "0.0.0.0/0"
          ip_protocol = "-1"
          description = "Allow all outbound"
        }
      ]
    }
  ]
}

resource "aws_elasticache_subnet_group" "cache" {
  subnet_ids = var.subnet_ids
  name       = "${local.prefix}-redis-group"
}


resource "aws_elasticache_parameter_group" "cluster" {
  for_each = local.cluster_services

  name   = "${local.prefix}-${each.key}-cluster-params"
  family = var.parameter_group_family

  parameter {
    name  = "cluster-enabled"
    value = "yes"
  }

  tags = merge(local.default_tags, 
           { Service = each.key } )
}

resource "aws_elasticache_replication_group" "cluster" {
  for_each = local.cluster_services

  replication_group_id = "${local.prefix}-${each.key}-cluster"
  description          = "Redis cluster for ${each.key}"

  engine               = "redis"
  engine_version       = var.engine_version
  node_type            = each.value.node_type
  port                 = 6379
  parameter_group_name = aws_elasticache_parameter_group.cluster[each.key].name

  num_node_groups         = each.value.num_shards
  replicas_per_node_group = each.value.replicas_per_shard

  subnet_group_name  = aws_elasticache_subnet_group.cache.id
  security_group_ids = module.sg.security_group_ids

  at_rest_encryption_enabled = true
  transit_encryption_enabled = false
  automatic_failover_enabled = each.value.num_shards > 1 ? true : false

  tags = merge(local.default_tags, {
    Service = each.key
    Role    = "cluster"
  })
}


resource "aws_elasticache_parameter_group" "standalone" {
  for_each = local.standalone_services

  name   = "${local.prefix}-${each.key}-standalone-params"
  family = var.parameter_group_family

  parameter {
    name  = "cluster-enabled"
    value = "no"
  }

  tags = merge(local.default_tags, { Service = each.key })
}

resource "aws_elasticache_replication_group" "standalone" {
  for_each = local.standalone_services

  replication_group_id = "${local.prefix}-${each.key}-standalone"
  description          = "Redis standalone for ${each.key}"

  engine               = "redis"
  engine_version       = var.engine_version
  node_type            = each.value.node_type
  port                 = 6379
  parameter_group_name = aws_elasticache_parameter_group.standalone[each.key].name

  num_node_groups         = 1
  replicas_per_node_group = 1

  subnet_group_name  = aws_elasticache_subnet_group.cache.id
  security_group_ids = module.sg.security_group_ids

  at_rest_encryption_enabled = true
  transit_encryption_enabled = false
  automatic_failover_enabled = true

  tags = merge(local.default_tags, {
    Service = each.key
    Role    = "standalone"
  })
}


resource "aws_secretsmanager_secret" "cache_credentials" {
  for_each = local.services_map

  name                    = "${local.prefix}/${each.key}/cache-credentials"
  recovery_window_in_days = 0

  tags = merge(local.default_tags, { Service = each.key })
}

resource "aws_secretsmanager_secret_version" "cache_credentials" {
  for_each = local.services_map

  secret_id = aws_secretsmanager_secret.cache_credentials[each.key].id
  secret_string = jsonencode(merge(
    each.value.cluster_mode ? {
      cluster_endpoint = aws_elasticache_replication_group.cluster[each.key].configuration_endpoint_address
      cluster_port     = "6379"
      cluster_password = ""
    } : {},
    each.value.standalone_mode ? {
      standalone_endpoint = aws_elasticache_replication_group.standalone[each.key].primary_endpoint_address
      standalone_port     = "6379"
      standalone_password = ""
    } : {}
  ))
}
