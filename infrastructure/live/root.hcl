locals {
  project = "DMB"
  region  = "eu-west-1"
  env = basename(get_terragrunt_dir())
}

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

