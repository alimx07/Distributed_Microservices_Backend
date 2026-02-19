include {
  path = find_in_parent_folders("root.hcl")
}


terraform {
    source = "../../../terrafrom_modules//EKS"
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
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}
EOF
}


inputs = {
    cluster_name = "DMB"
    subnet_ids = dependency.vpc.outputs.infra_subnets
}