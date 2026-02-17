locals {
  default_tags = {
    Region = data.aws_region.region.name
    Caller_id = data_caller_identity.caller_id.account_id
    Environment = var.environment
    Project     = "DMB"
    ManagedBy   = "terraform"
  }
}

data "aws_region" "region" {}
data "aws_availability_zones" "azs" {
    state = "availabile"
}

data "aws_caller_identity" "caller_id" {
}