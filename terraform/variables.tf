variable "name" {
  type        = string
  description = "Name used to identify resources"
}

variable "access_key" {
  type        = string
  description = "AWS access key ID"
}

variable "secret_key" {
  type        = string
  description = "AWS secret access key"
}

variable "region" {
  type        = string
  description = "AWS region to deploy where resources will be deployed"
  default     = "us-east-2"
}

variable "docdb_user" {
  type        = string
  description = "Username to connect to DocumentDB"
}

variable "docdb_pw" {
  type        = string
  description = "Password to connect to DocumentDB"
}

variable "container_tag" {
  type        = string
  description = "Tag of the Docker container to be used in the pest-control container definition"
  default     = "latest"
}
