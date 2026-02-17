locals {
  azs = length(var.azs) > 0 ? var.azs : data.aws_availability_zones.azs.names
  nat_gateways_num = var.one_nat_gateway ? 1 : var.one_nat_gateway_per_az ? length(local.azs) : 1
  default_tags = merge({
    Caller_id = data.aws_caller_identity.caller_id.account_id
  }, var.default_tags
  )
}

data "aws_region" "region" {}
data "aws_availability_zones" "azs" {
    state = "available"
}

data "aws_caller_identity" "caller_id" {
}
