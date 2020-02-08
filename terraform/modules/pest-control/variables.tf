variable "name" {
  type        = string
  description = "Name used to identify resources"
}

variable "container_tag" {
  type        = string
  description = "Tag of the pest-control container in the registry to be used"
  default     = "latest"
}

variable "cluster_id" {
  type        = string
  description = "ID of the ECS cluster that the pest-control service will run in"
}

variable "security_groups" {
  type        = list(string)
  description = "VPC security groups for the pest-control service load balancer"
}

variable "subnets" {
  type        = list(string)
  description = "VPC subnets for the pest-control service load balancer"
}

variable "db_endpoint" {
  type        = string
  description = "Endpoint of the DocumentDB cluster"
}

variable "db_port" {
  type        = string
  description = "Port that the DocumentDB cluster is listening on"
}

variable "db_user" {
  type        = string
  description = "Master username for the DocumentDB cluster"
}

variable "db_pw" {
  type        = string
  description = "Master password for the DocumentDB cluster"
}
