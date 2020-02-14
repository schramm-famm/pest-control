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

variable "db_host" {
  type        = string
  description = "Host of the database"
  default     = "localhost"
}

variable "db_port" {
  type        = string
  description = "Port that the database is listening on"
  default     = "27017"
}

variable "db_user" {
  type        = string
  description = "Username to connect to database"
}

variable "db_pw" {
  type        = string
  description = "Password to connect to database"
}

variable "container_tag" {
  type        = string
  description = "Tag of the Docker container to be used in the pest-control container definition"
  default     = "latest"
}
