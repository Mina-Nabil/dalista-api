output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "alb_dns_name" {
  description = "ALB DNS name"
  value       = aws_lb.main.dns_name
}

output "api_url" {
  description = "Full API URL"
  value       = "https://${var.api_subdomain}.${var.domain_name}"
}

output "ecr_repository_url" {
  description = "ECR repository URL"
  value       = data.aws_ecr_repository.main.repository_url
}

output "rds_endpoint" {
  description = "RDS endpoint"
  value       = aws_db_instance.main.address
}

output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.main.name
}

output "ecs_service_name" {
  description = "ECS API service name"
  value       = aws_ecs_service.api.name
}

output "github_actions_role_arn" {
  description = "IAM role ARN for GitHub Actions OIDC (set as AWS_DEPLOY_ROLE_ARN secret)"
  value       = aws_iam_role.github_actions.arn
}

output "app_secret_arn" {
  description = "Secrets Manager ARN for app secrets"
  value       = aws_secretsmanager_secret.app.arn
}

output "private_subnet_ids" {
  description = "Private subnet IDs (used by the deploy workflow's RunTask)"
  value       = [aws_subnet.private_a.id, aws_subnet.private_b.id]
}

output "ecs_security_group_id" {
  description = "ECS security group ID (used by the deploy workflow's RunTask)"
  value       = aws_security_group.ecs.id
}
