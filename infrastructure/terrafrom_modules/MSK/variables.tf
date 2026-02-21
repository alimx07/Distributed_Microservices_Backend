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

variable "kafka_version" {
  type    = string
  default = "3.6.0"
}

variable "number_of_broker_nodes" {
  type    = number
  default = 3
}

variable "instance_type" {
  type    = string
  default = "kafka.t3.small"
}

variable "volume_size" {
  type    = number
  default = 20
}


variable "debezium_plugin_zip_path" {
  type        = string
}

# variable "db_primary_host" {
#   type        = string
#   sensitive = true
# }

variable "secret_arn" {
  type = string
  sensitive = true
}

# variable "db_connect_user" {
#   type    = string
#   sensitive = true
# }

# variable "db_connect_password" {
#   type      = string
#   sensitive = true
# }

# variable "db_name" {
#   type    = string
#   default = "postdb"
# }

variable "default_tags" {
  type    = map(string)
  default = {}
}


variable "bucket_name" {
  type = string
}

# variable "bucket_arn" {
#   type = string
# }