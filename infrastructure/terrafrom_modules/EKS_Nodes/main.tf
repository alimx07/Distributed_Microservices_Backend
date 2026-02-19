  terraform {
    backend "s3" {}
  }

resource "aws_eks_node_group" "this" {
    cluster_name = var.cluster_name 
    node_group_name = "${local.prefix}-X"
    node_role_arn =  aws_iam_role.eks_node_role.arn
    subnet_ids = var.subnet_ids

    instance_types = ["t3.small"]

    scaling_config {
        desired_size = var.desired_size
        min_size = var.min_size
        max_size = var.max_size
    }

  #   # ignore changes happen by auto-scaler
    lifecycle {
    ignore_changes = [scaling_config[0].desired_size]
  }

  tags = local.default_tags

  depends_on = [ aws_iam_role_policy_attachment.cni_policy , aws_iam_role_policy_attachment.ecr_policy , aws_iam_role_policy_attachment.worker_node_policy]
}

resource "aws_iam_role" "eks_node_role" {
  name = "eks-node-roletest"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Effect = "Allow",
      Principal = { Service = "ec2.amazonaws.com" },
      Action = "sts:AssumeRole"
    }]
  })
  tags = local.default_tags
}

resource "aws_iam_role_policy_attachment" "worker_node_policy" {
  role       = aws_iam_role.eks_node_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
}

resource "aws_iam_role_policy_attachment" "cni_policy" {
  role       = aws_iam_role.eks_node_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
}

resource "aws_iam_role_policy_attachment" "ecr_policy" {
  role       = aws_iam_role.eks_node_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

# addons

resource "aws_eks_addon" "coreDns" {
    addon_name = "coredns"
    cluster_name = var.cluster_name
    addon_version = "v1.13.2-eksbuild.1"

    depends_on = [ aws_eks_node_group.this ]
}


# Route53 record
# Assume Zone already created
resource "aws_route53_record" "ingress" {
  zone_id = data.aws_route53_zone.this.zone_id
  name    = "*.${var.environment}.${var.zone_url}"
  type    = "A"

  alias {
    name                   = data.aws_lb.nginx_nlb.dns_name
    zone_id                = data.aws_lb.nginx_nlb.zone_id
    evaluate_target_health = false
  }
}

data "aws_route53_zone" "this" {
  name = var.zone_url
}

data "aws_lb" "nginx_nlb" {
  tags = {
    "kubernetes.io/service-name" = "ingress-nginx-ns/ingress-nginx-controller"
  }

  depends_on = [helm_release.nginx_ingress]
}

