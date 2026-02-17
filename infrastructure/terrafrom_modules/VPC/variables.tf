variable "cidr_block" {
    type = string
}

variable "enable_dns_hostnames" {
    type = bool
    default = false
}

variable "enable_dns_support" {
    type = bool
    default = true
}

variable "azs" {
    type = list(string)
    default = []
}

variable "map_public_ip_on_launch" {
    type = bool
    default = true
}

variable "cidr_subnet_mask" {
    type = number
    default = 8
}

variable "one_nat_gateway" {
    type = bool
    default = false
}

variable "one_nat_gateway_per_az" {
    type = bool
    default = false
}

variable "environment" {
    type = string
}

variable "default_tags" {
    type = map(string)
}