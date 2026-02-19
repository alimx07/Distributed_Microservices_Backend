locals {
  perfix = "tf-dmb-${var.environment}"
  oidc_issuer_url = aws_eks_cluster.this.identity[0]["oidc"][0]["issuer"]
  oidc_provider   = replace(local.oidc_issuer_url, "https://", "")
  oidc_provider_arn = aws_iam_openid_connect_provider.eks.arn
    default_tags = merge({
    Caller_id = data.aws_caller_identity.caller_id.account_id
    Name = "${local.perfix}-eks"
  }, var.default_tags
  )
  
}

data "aws_region" "region" {}
data "aws_availability_zones" "azs" {
    state = "available"
}

data "aws_caller_identity" "caller_id" {
}

