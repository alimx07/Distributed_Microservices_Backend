include {
  path = find_in_parent_folders("root.hcl")
}

locals {
  env = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
}

terraform {
  source = "../../../terrafrom_modules//ElasticCache"
}

dependency "vpc" {
  config_path = "../vpc"
  mock_outputs = {
    vpc_id          = "vpc-mock"
    private_subnets = ["subnet-mock-1", "subnet-mock-2", "subnet-mock-3"]
  }
}

inputs = {
  environment  = local.env.environment
  vpc_id       = dependency.vpc.outputs.vpc_id
  subnet_ids   = dependency.vpc.outputs.private_subnets
  private_cidr = "10.0.0.0/16"

  node_type              = "cache.t3.micro"
  engine_version         = "7.1"
  parameter_group_family = "redis7"

  # Cluster mode: 3 shards x 1 replica = 6 nodes (mirrors docker-compose)
  num_shards         = 3
  replicas_per_shard = 1

  default_tags = local.env.default_tags
}
