# ------------------------------------------------------------------------------
# ECS task execution role — pulls images and injects secrets at task start
# ------------------------------------------------------------------------------

resource "aws_iam_role" "ecs_execution" {
  name = "dalista-${var.environment}-ecs-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_execution_base" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy" "ecs_execution_secrets" {
  name = "dalista-${var.environment}-secrets-access"
  role = aws_iam_role.ecs_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["secretsmanager:GetSecretValue"]
      Resource = [
        aws_secretsmanager_secret.app.arn,
        # RDS-managed master password secret.
        "arn:aws:secretsmanager:${var.aws_region}:*:secret:rds!*"
      ]
    }]
  })
}

# ------------------------------------------------------------------------------
# ECS task role — the app makes no direct AWS API calls yet (minimal trust).
# ------------------------------------------------------------------------------

resource "aws_iam_role" "ecs_task" {
  name = "dalista-${var.environment}-ecs-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

# ------------------------------------------------------------------------------
# GitHub Actions OIDC provider + deployer role
#
# The OIDC provider is account-global (one per issuer), managed only in the
# default/staging workspace. Production references it via the constructed ARN.
# ------------------------------------------------------------------------------

resource "aws_iam_openid_connect_provider" "github" {
  count           = terraform.workspace == "default" ? 1 : 0
  url             = "https://token.actions.githubusercontent.com"
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"]
}

locals {
  github_oidc_provider_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/token.actions.githubusercontent.com"

  # Accept both the ID-suffixed subject (immutable-claims orgs) and the plain
  # form, so an org-level toggle in GitHub doesn't silently break deploys. Both
  # stay pinned to this one repository.
  github_oidc_subjects = [
    "repo:${var.github_org}@${var.github_org_id}/${var.github_repo}@${var.github_repo_id}:*",
    "repo:${var.github_org}/${var.github_repo}:*",
  ]
}

# Workspace-scoped role name so staging and production coexist in IAM.
# default/staging → dalista-github-actions
# production      → dalista-production-github-actions
resource "aws_iam_role" "github_actions" {
  name = terraform.workspace == "default" ? "dalista-github-actions" : "dalista-${var.environment}-github-actions"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRoleWithWebIdentity"
      Principal = { Federated = local.github_oidc_provider_arn }
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
        }
        StringLike = {
          "token.actions.githubusercontent.com:sub" = local.github_oidc_subjects
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "github_actions" {
  name = "dalista-github-actions-deploy"
  role = aws_iam_role.github_actions.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ECR"
        Effect = "Allow"
        Action = [
          "ecr:GetAuthorizationToken",
          "ecr:BatchCheckLayerAvailability",
          "ecr:GetDownloadUrlForLayer",
          "ecr:BatchGetImage",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload"
        ]
        Resource = "*"
      },
      {
        Sid    = "ECS"
        Effect = "Allow"
        Action = [
          "ecs:UpdateService",
          "ecs:DescribeServices",
          "ecs:DescribeTasks",
          "ecs:RunTask",
          "ecs:StopTask",
          "ecs:DescribeTaskDefinition",
          "ecs:RegisterTaskDefinition",
          "ecs:DeregisterTaskDefinition",
          "ecs:ListTasks"
        ]
        Resource = "*"
      },
      {
        Sid      = "PassRole"
        Effect   = "Allow"
        Action   = "iam:PassRole"
        Resource = [aws_iam_role.ecs_execution.arn, aws_iam_role.ecs_task.arn]
      }
    ]
  })
}
