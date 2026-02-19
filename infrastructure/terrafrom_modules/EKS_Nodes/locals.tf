locals {
  prefix = "tf-dmp-${var.environment}"
  default_tags = merge({
    Caller_id = data.aws_caller_identity.caller_id.account_id
    Name = "${local.prefix}-nodegroup"
  },var.default_tags)
}

data "aws_region" "region" {}

data "aws_caller_identity" "caller_id" {
}