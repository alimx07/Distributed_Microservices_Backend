variable "environment" {
  type = string
}

variable "subnet_ids" {
  type        = list(string)
}

variable "vpc_id" {
  type        = string
}

variable "allowed_cidr_block" {
  type        = string
  default     = ""
}

variable "services" {
  type = list(object({
    name                       = string
    db_name                    = string
    enable_logical_replication = optional(bool, false)
  }))
}

variable "instance_class" {
  type    = string
  default = "db.t3.micro"
}

variable "allocated_storage" {
  type    = number
  default = 20
}

variable "engine_version" {
  type    = string
  default = "17"
}

variable "default_tags" {
  type    = map(string)
}
