variable "environment" {
    type = string
}

variable "cluster_name" {
    type = string
    default = "Cluster"
}

variable "subnet_ids" {
    type = list(string)
}

variable "region" {
    type = string
}

variable "default_tags" {
    type = map(string)
}