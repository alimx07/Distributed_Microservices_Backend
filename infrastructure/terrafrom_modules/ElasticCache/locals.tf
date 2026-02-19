locals {
  prefix = "tf-dmb-${var.environment}"

  services_map = {
    for svc in var.services : svc.name => svc
  }

  cluster_services = {
    for name, svc in local.services_map : name => svc if svc.cluster_mode
  }

  standalone_services = {
    for name, svc in local.services_map : name => svc if svc.standalone_mode
  }

  default_tags = merge(var.default_tags, {
    Name      = "${local.prefix}-elasticache"
    Caller_id = data.aws_caller_identity.caller_id.account_id
  })
}
