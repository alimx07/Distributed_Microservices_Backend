include {
  path = find_in_parent_folders("root.hcl")
}

locals {
  env = read_terragrunt_config(find_in_parent_folders("env.hcl")).locals
}

terraform {
    source = "../../../terrafrom_modules//VPC"
}

inputs = {
    environment = local.env.environment
    cidr_block = "10.0.0.0/16"   
    cidr_subnet_mask = 8
    enable_dns_hostnames = true
    one_nat_gateway = false
    one_nat_gateway_per_az = true
    default_tags = local.env.default_tags

}