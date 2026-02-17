variable "vpc_id" {
  type        = string
  description = "VPC ID where security groups will be created"
}

variable "security_groups" {
  type = list(object({
    name        = string
    description = optional(string, "Managed by Terraform")
    ingress_rules = optional(list(object({
      cidr_block  = string
      from_port   = optional(number , 0)
      to_port     = optional(number , 0)
      ip_protocol = optional(string , "-1")
      description = optional(string)
    })), [])
    egress_rules = optional(list(object({
      cidr_block  = string
      from_port   = optional(number , 0)
      to_port     = optional(number , 0)
      ip_protocol = optional(string , "-1")
      description = optional(string)
    })), [])
  }))
  description = "List of security groups with their ingress and egress rules"
  default     = []
}

variable "environment" {
    type = string
}
