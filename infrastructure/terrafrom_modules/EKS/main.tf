  terraform {
    backend "s3" {}
  }

resource "aws_eks_cluster" "this" {
    name = "${local.perfix}-EKS"
    role_arn = aws_iam_role.eks_role.arn
    version = "1.33"
    access_config {
    authentication_mode = "API_AND_CONFIG_MAP"
      }
    vpc_config {
      endpoint_public_access = true
      endpoint_private_access = true
      # infra subnets
      subnet_ids = var.subnet_ids
    }


  bootstrap_self_managed_addons = false
  tags = local.default_tags

  depends_on = [ aws_iam_role_policy_attachment.eks_policy]
}


resource "aws_iam_role" "eks_role" {
  name = "${local.perfix}-eks-role"

  # trust policy
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_policy" {
  role = aws_iam_role.eks_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
}



# Insert into local

resource "null_resource" "kubeconfig" {
  depends_on = [
    aws_eks_cluster.this
  ]

  triggers = {
    cluster_name = aws_eks_cluster.this.name
    endpoint     = aws_eks_cluster.this.endpoint
  }

  provisioner "local-exec" {
    command = <<EOT
    aws eks update-kubeconfig \
    --region ${var.region} \
    --name ${aws_eks_cluster.this.name}
EOT
  }
}



# Access Entry

resource "aws_eks_access_entry" "name" {
  cluster_name = aws_eks_cluster.this.name
  principal_arn = "arn:aws:iam::408502715955:role/TerraformRole_v1"
  tags = local.default_tags
}




resource "aws_eks_addon" "vpc_cni" {
    addon_name = "vpc-cni"
    cluster_name = aws_eks_cluster.this.name

    # aws eks describe-addon-versions     
    #--kubernetes-version=1.33     
    #--addon-name=vpc-cni     
    #--query='addons[].addonVersions[].addonVersion' 
    #--region eu-west-1

    addon_version = "v1.21.1-eksbuild.3"
    service_account_role_arn = module.vpc-cni.iam_role_arn
    tags = var.default_tags
    depends_on = [ aws_iam_openid_connect_provider.eks ]
}

resource "aws_eks_addon" "kube_proxy" {
    addon_name = "kube-proxy"
    cluster_name = aws_eks_cluster.this.name
    addon_version = "v1.33.7-eksbuild.2"
    depends_on = [ aws_eks_addon.vpc_cni ]
}



# OIDC Provider - Required for IRSA to work
data "tls_certificate" "eks" {
  url = aws_eks_cluster.this.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "eks" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.eks.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.this.identity[0].oidc[0].issuer
  tags            = local.default_tags
}
