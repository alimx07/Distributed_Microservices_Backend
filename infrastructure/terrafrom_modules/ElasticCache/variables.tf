variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "private_cidr" {
  type = string
}

variable "services" {
  type = list(object({
    name               = string
    cluster_mode       = optional(bool, false)
    standalone_mode    = optional(bool, false)
    num_shards         = optional(number, 3)
    replicas_per_shard = optional(number, 1)
    node_type          = optional(string, "cache.t3.micro")
  }))
}

variable "engine_version" {
  type    = string
  default = "7.1"
}

variable "parameter_group_family" {
  type    = string
  default = "redis7"
}

variable "default_tags" {
  type    = map(string)
  default = {}
}
