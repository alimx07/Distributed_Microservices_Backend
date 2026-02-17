locals {
  prefix = "tf-nodegroup-${var.environment}"
}

data "aws_region" "region" {}

data "aws_caller_identity" "caller_id" {
}