# ------------------------------------------------------------------------------
# ECR repository
#
# Repo names are account-unique and shared across environments (images are
# distinguished by `${env}-<sha>` tags), so the repo is managed only in the
# default/staging workspace and referenced elsewhere via the data source.
# ------------------------------------------------------------------------------

resource "aws_ecr_repository" "main" {
  count                = terraform.workspace == "default" ? 1 : 0
  name                 = "dalista-be"
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = { Name = "dalista-be" }
}

data "aws_ecr_repository" "main" {
  name       = "dalista-be"
  depends_on = [aws_ecr_repository.main]
}

resource "aws_ecr_lifecycle_policy" "main" {
  count      = terraform.workspace == "default" ? 1 : 0
  repository = aws_ecr_repository.main[0].name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Expire untagged images after 1 day"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 1
        }
        action = { type = "expire" }
      },
      {
        rulePriority = 2
        description  = "Keep last 10 tagged images"
        selection = {
          tagStatus     = "tagged"
          tagPrefixList = ["staging-", "production-"]
          countType     = "imageCountMoreThan"
          countNumber   = 10
        }
        action = { type = "expire" }
      }
    ]
  })
}
