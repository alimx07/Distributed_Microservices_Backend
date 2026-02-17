output "cluster_arn" {
    value = aws_eks_cluster.this.arn
}

output "cluster_name" {
    value = aws_eks_cluster.this.name
}

output "eks_cluster_id" {
    value = aws_eks_cluster.this.id
}

output "cluster_endpoint" {
    value = aws_eks_cluster.this.endpoint
}

output "eks_cert" {
    value = aws_eks_cluster.this.certificate_authority[0].data
}

output "eks_oidc" {
    value = local.oidc_provider_arn
}

# 01:31:44.168 STDOUT [eks] terraform: eks_oidc = tolist([
# 01:31:44.168 STDOUT [eks] terraform:   {
# 01:31:44.168 STDOUT [eks] terraform:     "oidc" = tolist([
# 01:31:44.168 STDOUT [eks] terraform:       {
# 01:31:44.168 STDOUT [eks] terraform:         "issuer" = "https://oidc.eks.eu-west-1.amazonaws.com/id/DD788FCAD9223AE42A32A20F9C57F9AD"
# 01:31:44.168 STDOUT [eks] terraform:       },
# 01:31:44.168 STDOUT [eks] terraform:     ])
# 01:31:44.168 STDOUT [eks] terraform:   },
# 01:31:44.168 STDOUT [eks] terraform: ])
