locals {
  prefix = "tf-dmb-${var.environment}"

  services_map = {
    for svc in var.services : svc.name => svc
  }

  logical_replication_services = {
    for name, svc in local.services_map : name => svc if svc.enable_logical_replication
  }

  default_tags = merge(var.default_tags, {
    Caller_id = data.aws_caller_identity.caller_id.account_id
    Name = "${local.prefix}-rds"
  })
}

data "aws_caller_identity" "caller_id" {
}

