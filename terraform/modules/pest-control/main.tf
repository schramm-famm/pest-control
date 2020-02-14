data "aws_region" "pest-control" {}

resource "aws_cloudwatch_log_group" "pest-control" {
  name = "${var.name}_pest-control"
}

resource "aws_ecs_task_definition" "pest-control" {
  family       = "${var.name}_pest-control"
  network_mode = "bridge"

  container_definitions = <<EOF
[
  {
    "name": "${var.name}_pest-control",
    "image": "343660461351.dkr.ecr.us-east-2.amazonaws.com/pest-control:${var.container_tag}",
    "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
            "awslogs-group": "${aws_cloudwatch_log_group.pest-control.name}",
            "awslogs-region": "${data.aws_region.pest-control.name}",
            "awslogs-stream-prefix": "${var.name}"
        }
    },
    "cpu": 10,
    "memory": 128,
    "environment": [
        {
            "name": "PESTCONTROL_DB_HOST",
            "value": "${var.db_host}"
        },
        {
            "name": "PESTCONTROL_DB_PORT",
            "value": "${var.db_port}"
        },
        {
            "name": "PESTCONTROL_DB_USER",
            "value": "${var.db_user}"
        },
        {
            "name": "PESTCONTROL_DB_PW",
            "value": "${var.db_pw}"
        }
    ],
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 80,
        "protocol": "tcp"
      }
    ]
  }
]
EOF
}

resource "aws_elb" "pest-control" {
  name            = "${var.name}-pest-control"
  subnets         = var.subnets
  security_groups = var.security_groups

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_ecs_service" "pest-control" {
  name            = "${var.name}_pest-control"
  cluster         = var.cluster_id
  task_definition = aws_ecs_task_definition.pest-control.arn

  load_balancer {
    elb_name       = aws_elb.pest-control.name
    container_name = "${var.name}_pest-control"
    container_port = 80
  }

  desired_count = 1
}
