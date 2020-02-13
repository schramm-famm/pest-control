data "aws_region" "mongodb" {}

resource "aws_cloudwatch_log_group" "mongodb" {
  name = "${var.name}_mongodb"
}

resource "aws_ecs_task_definition" "mongodb" {
  family       = "${var.name}_mongodb"
  network_mode = "bridge"

  container_definitions = <<EOF
[
  {
    "name": "${var.name}_mongodb",
    "image": "mongo",
    "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
            "awslogs-create-group": "true",
            "awslogs-group": "${aws_cloudwatch_log_group.mongodb.name}",
            "awslogs-region": "${data.aws_region.mongodb.name}",
            "awslogs-stream-prefix": "${var.name}"
        }
    },
    "cpu": 10,
    "memory": 128,
    "environment": [
        {
            "name": "MONGO_INITDB_ROOT_USERNAME",
            "value": "${var.db_user}"
        },
        {
            "name": "MONGO_INITDB_ROOT_PASSWORD",
            "value": "${var.db_pw}"
        }
    ],
    "portMappings": [
      {
        "containerPort": 27017,
        "hostPort": 27017,
        "protocol": "tcp"
      }
    ]
  }
]
EOF
}

resource "aws_elb" "mongodb" {
  name            = "${var.name}-mongodb"
  subnets         = var.subnets
  security_groups = var.security_groups

  listener {
    instance_port     = 27017
    instance_protocol = "tcp"
    lb_port           = 27017
    lb_protocol       = "tcp"
  }
}

resource "aws_ecs_service" "mongodb" {
  name            = "${var.name}_mongodb"
  cluster         = var.cluster_id
  task_definition = aws_ecs_task_definition.mongodb.arn

  load_balancer {
    elb_name       = aws_elb.mongodb.name
    container_name = "${var.name}_mongodb"
    container_port = 27017
  }

  desired_count = 1
}
