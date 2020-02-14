provider "aws" {
  access_key = var.access_key
  secret_key = var.secret_key
  region     = var.region
}

module "ecs_base" {
  source = "github.com/schramm-famm/bespin//modules/ecs_base"
  name   = var.name
}

resource "aws_security_group" "pest-control" {
  name        = "${var.name}_allow_testing"
  description = "Allow traffic necessary for integration testing"
  vpc_id      = module.ecs_base.vpc_id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 27017
    to_port     = 27017
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

module "ecs_cluster" {
  source                  = "github.com/schramm-famm/bespin//modules/ecs_cluster"
  name                    = var.name
  security_group_ids      = [aws_security_group.pest-control.id]
  subnets                 = module.ecs_base.vpc_public_subnets
  ec2_instance_profile_id = module.ecs_base.ecs_instance_profile_id
}

module "mongodb" {
  source          = "./modules/mongodb"
  name            = var.name
  cluster_id      = module.ecs_cluster.cluster_id
  security_groups = [aws_security_group.pest-control.id]
  subnets         = module.ecs_base.vpc_public_subnets
  db_user         = var.db_user
  db_pw           = var.db_pw
}

module "pest-control" {
  source          = "./modules/pest-control"
  name            = var.name
  container_tag   = var.container_tag
  cluster_id      = module.ecs_cluster.cluster_id
  security_groups = [aws_security_group.pest-control.id]
  subnets         = module.ecs_base.vpc_public_subnets
  db_host         = module.mongodb.host
  db_port         = "27017"
  db_user         = var.db_user
  db_pw           = var.db_pw
}
