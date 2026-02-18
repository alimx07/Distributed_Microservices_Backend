include {
  path = find_in_parent_folders("root.hcl")
}

locals {
  env = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
}

terraform {
  source = "../../../terrafrom_modules//RDS"
}

dependency "vpc" {
  config_path = "../vpc"
  # mock_outputs = {
  #   vpc_id          = "vpc-mock"
  #   private_subnets = ["subnet-mock-1", "subnet-mock-2"]
  # }
}

dependency "eks" {
  config_path = "../eks"
  # mock_outputs = {
  #   cluster_security_group_id = "sg-mock"
  # }
}

inputs = {
  environment = local.env.environment
  vpc_id      = dependency.vpc.outputs.vpc_id
  subnet_ids  = dependency.vpc.outputs.private_subnets
  allowed_cidr_block = "10.0.0.0/16"
  services = [
    {
      name    = "userx"
      db_name = "userdb"
    },
    {
      name                       = "post"
      db_name                    = "postdb"
      enable_logical_replication = true
    },
    {
      name    = "follow"
      db_name = "followdb"
    }
  ]

  instance_class    = "db.t3.micro"
  allocated_storage = 20
  engine_version    = "17"
  default_tags      = local.env.default_tags
}
