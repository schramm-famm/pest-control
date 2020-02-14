variable "name" {
  type        = string
  description = "Name used to identify resources"
}

variable "cluster_id" {
  type        = string
  description = "ID of the ECS cluster that the mongodb service will run in"
}

variable "security_groups" {
  type        = list(string)
  description = "VPC security groups for the mongodb service load balancer"
}

variable "subnets" {
  type        = list(string)
  description = "VPC subnets for the mongodb service load balancer"
}

variable "db_user" {
  type        = string
  description = "Master username for the MongoDB database"
}

variable "db_pw" {
  type        = string
  description = "Master password for the MongoDB database"
}
