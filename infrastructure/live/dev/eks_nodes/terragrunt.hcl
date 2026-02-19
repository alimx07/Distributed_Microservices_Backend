include {
  path = find_in_parent_folders("root.hcl")
}


terraform {
    source = "../../../terrafrom_modules//EKS_Nodes"
}

dependency "eks" {
  config_path = "../eks"
}

dependency "vpc" {
    config_path = "../vpc"
}

generate "provider_extra" {
  path      = "provider_extra.tf"
  if_exists = "overwrite"
  contents  = <<EOF
  

terraform {
  required_providers {
    kubectl = {
      source  = "alekc/kubectl"
      version = ">= 2.0"
    }
  }
}
provider "kubernetes" {
  host                   = "${dependency.eks.outputs.cluster_endpoint}"
  cluster_ca_certificate = base64decode("${dependency.eks.outputs.eks_cert}")

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args = ["eks", "get-token", "--cluster-name", "${dependency.eks.outputs.cluster_name}"]
  }
}

provider "helm" {
  kubernetes = {
    host                   = "${dependency.eks.outputs.cluster_endpoint}"
    cluster_ca_certificate = base64decode("${dependency.eks.outputs.eks_cert}")
    exec = {
      api_version = "client.authentication.k8s.io/v1beta1"
      args        = ["eks", "get-token", "--cluster-name", "${dependency.eks.outputs.cluster_name}"]
      command     = "aws"
    }
  }
}

provider "kubectl" {
  host                   = "${dependency.eks.outputs.cluster_endpoint}"
  cluster_ca_certificate = base64decode("${dependency.eks.outputs.eks_cert}")
  load_config_file       = false

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args = ["eks", "get-token", "--cluster-name", "${dependency.eks.outputs.cluster_name}"]
  }
}
EOF
}


inputs = {
    cluster_name = dependency.eks.outputs.cluster_name
    subnet_ids = dependency.vpc.outputs.private_subnets
    max_size = 5
    min_size = 2
    desired_size = 4
    eks_oidc = dependency.eks.outputs.eks_oidc
}