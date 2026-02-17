include {
  path = find_in_parent_folders("root.hcl")
}

locals {
  env = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
}

terraform {
    source = "../../../terrafrom_modules//EKS"
}

dependency "vpc" {
  config_path = "../vpc"
    mock_outputs = {
          infra_subnets = ["mock-vpc-output"]
    }

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
    environment = local.env.environment
    cluster_name = "DMB"
    subnet_ids = dependency.vpc.outputs.infra_subnets
    region = local.env.region
    default_tags = local.env.default_tags
}