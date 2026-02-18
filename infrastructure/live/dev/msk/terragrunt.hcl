include {
  path = find_in_parent_folders("root.hcl")
}

locals {
  env = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
}

terraform {
  source = "../../../terrafrom_modules//MSK"
}

dependency "vpc" {
  config_path = "../vpc"

}

dependency "rds" {
  config_path = "../rds"

}

inputs = {
  environment = local.env.environment
  vpc_id      = dependency.vpc.outputs.vpc_id
  subnet_ids  = dependency.vpc.outputs.private_subnets
  private_cidr = "10.0.0.0/16"

  kafka_version          = "4.0.x.kraft"
  number_of_broker_nodes = 3
  instance_type          = "kafka.m7g.large"
  volume_size        = 20

  
  debezium_plugin_zip_path = "/home/ali-mohamed/projects/DMB/debezium-connector-postgres-3.4.1.Final-plugin.tar.gz"
  db_primary_host          = dependency.rds.outputs.db_connections["post"].primary_host
  db_connect_user          = "logical_rep"
  db_connect_password      = "1234"
  db_name                  = "postdb"

  default_tags = local.env.default_tags

  bucket_name = "amx-bucket-724"
  # bucket_arn = "arn:aws:s3:::amx-bucket-724"
}
