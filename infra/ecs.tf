# ------------------------------------------------------------------------------
# ECS cluster
# ------------------------------------------------------------------------------

resource "aws_ecs_cluster" "main" {
  name = "dalista-${var.environment}"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = { Name = "dalista-${var.environment}-cluster" }
}

# ------------------------------------------------------------------------------
# Shared task config
# ------------------------------------------------------------------------------

locals {
  # Secrets injected from the Secrets Manager JSON blob by key name (:KEY::).
  task_secrets = [
    for key in [
      "JWT_KEY",
      "DB_HOST",
      "DB_PORT",
      "DB_USER",
      "DB_PASSWORD",
      "DB_NAME",
      "SENTRY_DSN",
      "GOOSE_DBSTRING",
      ] : {
      name      = key
      valueFrom = "${aws_secretsmanager_secret.app.arn}:${key}::"
    }
  ]

  # Non-sensitive environment variables.
  task_environment = [
    { name = "APP_ENV", value = var.environment },
    { name = "FRONT_END_URL", value = trimspace(var.front_end_url) },
    { name = "CORS_ALLOW_ALL", value = "false" },
    { name = "JWT_EXPIRY_HOURS", value = tostring(var.jwt_expiry_hours) },
    { name = "RATE_LIMIT_PER_MINUTE", value = tostring(var.rate_limit_per_minute) },
    { name = "DB_SSLMODE", value = "require" },
    { name = "GOOSE_DRIVER", value = "postgres" },
    { name = "GOOSE_MIGRATION_DIR", value = "/app/migrations" },
  ]

  # Deploy workflow pushes a fresh ${env}-<sha> tag per commit and points the
  # task definition at it via RegisterTaskDefinition; this baseline uses the
  # env-tagged image for the initial `terraform apply`.
  ecr_image = "${data.aws_ecr_repository.main.repository_url}:${var.environment}-latest"
}

# ------------------------------------------------------------------------------
# API server task definition + service
# ------------------------------------------------------------------------------

resource "aws_ecs_task_definition" "api" {
  family                   = "dalista-${var.environment}-api"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name      = "api"
    image     = local.ecr_image
    essential = true

    portMappings = [{
      containerPort = 8080
      protocol      = "tcp"
    }]

    environment = local.task_environment
    secrets     = local.task_secrets

    # Liveness is checked by the ALB target group (GET /health). The distroless
    # runtime image has no shell for a container-level health command.

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "api"
      }
    }
  }])

  tags = { Name = "dalista-${var.environment}-api" }
}

resource "aws_ecs_service" "api" {
  name            = "dalista-${var.environment}-api"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = [aws_subnet.private_a.id, aws_subnet.private_b.id]
    security_groups  = [aws_security_group.ecs.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "api"
    container_port   = 8080
  }

  deployment_minimum_healthy_percent = 100
  deployment_maximum_percent         = 200

  # The deploy workflow swaps the image via force-new-deployment; ignore the
  # task_definition drift that introduces so terraform plans stay clean.
  lifecycle {
    ignore_changes = [task_definition]
  }

  depends_on = [aws_lb_listener.https]

  tags = { Name = "dalista-${var.environment}-api-service" }
}

# ------------------------------------------------------------------------------
# Migration task definition (standalone RunTask, invoked by the deploy workflow)
# ------------------------------------------------------------------------------

resource "aws_ecs_task_definition" "migrate" {
  family                   = "dalista-${var.environment}-migrate"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name      = "migrate"
    image     = local.ecr_image
    essential = true

    # goose reads GOOSE_DRIVER and GOOSE_DBSTRING from the environment/secrets.
    command = ["/goose", "-dir", "/app/migrations", "up"]

    environment = local.task_environment
    secrets     = local.task_secrets

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "migrate"
      }
    }
  }])

  tags = { Name = "dalista-${var.environment}-migrate" }
}

# ------------------------------------------------------------------------------
# Sync-permissions task definition (standalone RunTask, after migrate on deploy)
# ------------------------------------------------------------------------------

resource "aws_ecs_task_definition" "syncperms" {
  family                   = "dalista-${var.environment}-syncperms"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name      = "syncperms"
    image     = local.ecr_image
    essential = true

    command = ["/syncpermissions"]

    environment = local.task_environment
    secrets     = local.task_secrets

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "syncperms"
      }
    }
  }])

  tags = { Name = "dalista-${var.environment}-syncperms" }
}
