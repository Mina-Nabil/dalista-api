# ------------------------------------------------------------------------------
# CloudWatch log group for ECS task logs (api + migrate streams)
# ------------------------------------------------------------------------------

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/dalista-${var.environment}"
  retention_in_days = 30

  tags = { Name = "dalista-${var.environment}-ecs-logs" }
}
