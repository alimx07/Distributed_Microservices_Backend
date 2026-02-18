variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnet IDs for broker nodes (count must match number_of_broker_nodes)"
}

variable "private_cidr" {
  type        = string
  description = "VPC private CIDR for security group ingress"
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
  description = "Local path to the Debezium PostgreSQL connector ZIP"
}

variable "db_primary_host" {
  type        = string
  description = "RDS primary endpoint for Debezium"
}

variable "db_connect_user" {
  type    = string
  default = "logical_rep"
}

variable "db_connect_password" {
  type      = string
  sensitive = true
}

variable "db_name" {
  type    = string
  default = "postdb"
}

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