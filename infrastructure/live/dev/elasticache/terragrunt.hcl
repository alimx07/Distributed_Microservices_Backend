include {
  path = find_in_parent_folders("root.hcl")
}

terraform {
  source = "../../../terrafrom_modules//ElasticCache"
}

dependency "vpc" {
  config_path = "../vpc"
}

inputs = {
  vpc_id       = dependency.vpc.outputs.vpc_id
  subnet_ids   = dependency.vpc.outputs.private_subnets
  private_cidr = "10.0.0.0/16"

  engine_version         = "7.1"
  parameter_group_family = "redis7"

  services = [
    {
      name               = "api-gateway"
      cluster_mode       = true
      standalone_mode    = true
      num_shards         = 3
      replicas_per_shard = 1
      node_type          = "cache.t3.micro"
    },
    {
      name               = "post"
      cluster_mode       = true
      standalone_mode    = false
      num_shards         = 3
      replicas_per_shard = 1
      node_type          = "cache.t3.micro"
    },
    {
      name               = "feed"
      cluster_mode       = true
      standalone_mode    = false
      num_shards         = 3
      replicas_per_shard = 1
      node_type          = "cache.t3.micro"
    },
    {
      name               = "userx"
      cluster_mode       = true
      standalone_mode    = true
      num_shards         = 3
      replicas_per_shard = 1
      node_type          = "cache.t3.micro"
    }
  ]
}
