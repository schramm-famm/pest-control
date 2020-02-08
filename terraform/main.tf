provider "aws" {
  access_key = var.access_key
  secret_key = var.secret_key
  region     = var.region
}

module "ecs_base" {
  source = "github.com/schramm-famm/bespin//modules/ecs_base"
  name   = var.name
}

module "ecs_cluster" {
  source                  = "github.com/schramm-famm/bespin//modules/ecs_cluster"
  name                    = var.name
  security_group_id       = module.ecs_base.vpc_default_security_group_id
  subnets                 = module.ecs_base.vpc_public_subnets
  ec2_instance_profile_id = module.ecs_base.ecs_instance_profile_id
}

resource "aws_docdb_subnet_group" "subnet_group" {
  name       = "${var.name}_subnet_group"
  subnet_ids = module.ecs_base.vpc_public_subnets
}

resource "aws_docdb_cluster_instance" "cluster_instances" {
  count              = 2
  identifier         = "${aws_docdb_cluster.docdb.id}-${count.index}"
  cluster_identifier = aws_docdb_cluster.docdb.id
  instance_class     = "db.r5.large"
}

resource "aws_docdb_cluster" "docdb" {
  cluster_identifier   = "${var.name}-docdb-cluster"
  db_subnet_group_name = aws_docdb_subnet_group.subnet_group.name
  master_username      = var.docdb_user
  master_password      = var.docdb_pw
  skip_final_snapshot  = true
}

module "pest-control" {
  source          = "./modules/pest-control"
  name            = var.name
  container_tag   = "1.0.0"
  cluster_id      = module.ecs_cluster.cluster_id
  security_groups = [module.ecs_base.vpc_default_security_group_id]
  subnets         = module.ecs_base.vpc_public_subnets
  db_endpoint     = aws_docdb_cluster.docdb.endpoint
  db_port         = aws_docdb_cluster.docdb.port
  db_user         = aws_docdb_cluster.docdb.master_username
  db_pw           = aws_docdb_cluster.docdb.master_password
}
