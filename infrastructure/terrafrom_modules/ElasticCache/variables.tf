variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "subnet_ids" {
  type        = list(string)
}

variable "private_cidr" {
  type        = string
}

variable "node_type" {
  type    = string
  default = "cache.t3.micro"
}

variable "engine_version" {
  type    = string
  default = "7.1"
}

variable "parameter_group_family" {
  type    = string
  default = "redis7"
}

variable "num_shards" {
  type    = number
  default = 3
}

variable "replicas_per_shard" {
  type    = number
  default = 1
}

variable "default_tags" {
  type    = map(string)
  default = {}
}


