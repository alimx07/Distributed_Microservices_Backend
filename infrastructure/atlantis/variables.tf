variable "gh_user" {
  type        = string
}

variable "gh_token" {
  type        = string
  sensitive   = true
}

variable "gh_webhook_secret" {
  type        = string
  sensitive   = true
}


variable "vpc_id" {
  type = string
}

variable "key_name" {
  description = "SSH_KEY"
  type        = string
  default     = ""
}

variable "terraform_version" {
  type        = string
  default     = "1.14.8"
}

variable "terragrunt_version" {
  type        = string
  default     = "1.0.0"
}

variable "terragrunt_atlantis_config_version" {
  type        = string
  default     = "1.21.1"
}

variable "atlantis_version" {
  type        = string
  default     = "0.41.0"
}
