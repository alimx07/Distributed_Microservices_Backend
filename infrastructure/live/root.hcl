locals {
  env_vars = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
  environment = local.env_vars.environment
  # project = "DMB"
  region = local.env_vars.region # easy access
  default_tags = {
    Environment = local.environment
    Project     = "DMB"
    ManagedBy   = "terraform"
  }
}

# Shared between All ENVs

remote_state {
  backend = "s3"

  config = {
    bucket         = "amx-bucket-724"
    key            = "${path_relative_to_include()}/terraform.tfstate"
    region         = local.region
    use_lockfile   = true
  }
}

generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite"
  contents  = <<EOF
provider "aws" {
  region = "${local.region}"
    assume_role {
    role_arn     = "arn:aws:iam::408502715955:role/TerraformRole_v1"
  }
}
EOF
}


inputs = local
