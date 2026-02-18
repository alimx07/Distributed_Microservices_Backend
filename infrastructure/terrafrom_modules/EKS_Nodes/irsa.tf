module "cert_manager_irsa" {
  source                = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version               = "~> 5.0"
  role_name             = "${local.prefix}-cert_manager_irsa"

  attach_cert_manager_policy = true

  # cert_manager_hosted_zone_arns = [data.aws_route53_zone.this.arn]
  oidc_providers = {
    ex = {
      provider_arn               = var.eks_oidc
      namespace_service_accounts = ["cert-manager-ns:cert-manager"]
    }
  }

  tags = merge(var.default_tags,
    {
      Name = "${local.prefix}-cert_manager_irsa"
    })
}


module "external_secrets_irsa" {
  source                = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version               = "~> 5.0"
  role_name             = "${local.prefix}-external-secrets-irsa"
  attach_external_secrets_policy = true

  oidc_providers = {
    ex = {
      provider_arn               = var.eks_oidc
      namespace_service_accounts = ["external-secrets-ns:external-secrets"]
    }
  }
    tags = merge(var.default_tags,
    {
      Name = "${local.prefix}-external-secrets-irsa"
    })
}