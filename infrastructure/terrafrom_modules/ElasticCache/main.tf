terraform {
  backend "s3" {}
}

locals {
  prefix = "tf-dmb-${var.environment}"

  default_tags = merge(var.default_tags, {
    Name = "${local.prefix}-elasticache"
  })
}

module "sg" {
  source      = "../SG"
  vpc_id      = var.vpc_id
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


resource "aws_elasticache_replication_group" "cluster" {
  replication_group_id = "${local.prefix}-redis-cluster"
  description          = "Redis cluster for DMB services"

  engine               = "redis"
  engine_version       = var.engine_version
  node_type            = var.node_type
  port                 = 6379
  parameter_group_name = aws_elasticache_parameter_group.cluster.name

  num_node_groups         = var.num_shards
  replicas_per_node_group = var.replicas_per_shard

  subnet_group_name  = aws_elasticache_subnet_group.cache.id
  security_group_ids = module.sg.security_group_ids

  at_rest_encryption_enabled = true
  transit_encryption_enabled = false
  automatic_failover_enabled = var.num_shards > 1 ? true : false

  tags = merge(local.default_tags, {
    Role = "cluster"
  })
}

resource "aws_elasticache_parameter_group" "cluster" {
  name   = "${local.prefix}-redis-cluster-params"
  family = var.parameter_group_family

  parameter {
    name  = "cluster-enabled"
    value = "yes"
  }

  tags = local.default_tags
}


resource "aws_elasticache_replication_group" "standalone" {
  replication_group_id = "${local.prefix}-redis-standalone"
  description          = "Redis standalone for user cache"

  engine               = "redis"
  engine_version       = var.engine_version
  node_type            = var.node_type
  port                 = 6379
  parameter_group_name = aws_elasticache_parameter_group.standalone.name

  num_node_groups         = 1
  replicas_per_node_group = 1

  subnet_group_name  = aws_elasticache_subnet_group.cache.id
  security_group_ids = module.sg.security_group_ids

  at_rest_encryption_enabled = true
  transit_encryption_enabled = false
  automatic_failover_enabled = true

  tags = merge(local.default_tags, {
    Role = "standalone"
  })
}

resource "aws_elasticache_parameter_group" "standalone" {
  name   = "${local.prefix}-redis-standalone-params"
  family = var.parameter_group_family

  parameter {
    name  = "cluster-enabled"
    value = "no"
  }

  tags = local.default_tags
}