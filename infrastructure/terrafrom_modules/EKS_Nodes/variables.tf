variable "cluster_name" {
    type = string
}

variable "subnet_ids" {
    type = list(string)
}

variable "max_size" {
    type = number
}
variable "min_size" {
    type = number
}

variable "desired_size" {
    type = number
}

variable "environment" {
    type = string
}


variable "eks_oidc" {
    type = string
}

variable "email" {
    type = string
    default = ""
}


variable "zone_url" {
    type = string
    default = "alimx07.com"
}

variable "default_tags" {
    type = map(string)
}