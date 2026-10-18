# ------------------------------------------------------------------------------
# Remote-state bootstrap (DAL-10)
#
# Creates the S3 bucket and DynamoDB lock table that back the main config's
# S3 remote state. This is a chicken-and-egg problem — the backend can't store
# its own backing store — so this module uses LOCAL state and is applied ONCE,
# before `terraform init` in the parent ../ directory.
#
#   cd infra/bootstrap
#   terraform init
#   terraform apply
#
# Its own terraform.tfstate is small and can be committed or kept locally; it
# changes only when this bucket/table definition changes.
# ------------------------------------------------------------------------------

terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = "dalista"
      ManagedBy = "terraform"
      Component = "tf-state-bootstrap"
    }
  }
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "state_bucket_name" {
  description = "S3 bucket for Terraform remote state (must match the backend block in ../main.tf)"
  type        = string
  default     = "dalista-terraform-state"
}

variable "lock_table_name" {
  description = "DynamoDB table for state locking (must match the backend block in ../main.tf)"
  type        = string
  default     = "dalista-terraform-locks"
}

resource "aws_s3_bucket" "state" {
  bucket = var.state_bucket_name

  # State is irreplaceable — guard against accidental `terraform destroy`.
  lifecycle {
    prevent_destroy = true
  }

  tags = { Name = var.state_bucket_name }
}

resource "aws_s3_bucket_versioning" "state" {
  bucket = aws_s3_bucket.state.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "state" {
  bucket = aws_s3_bucket.state.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "state" {
  bucket                  = aws_s3_bucket.state.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_dynamodb_table" "locks" {
  name         = var.lock_table_name
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  tags = { Name = var.lock_table_name }
}

output "state_bucket" {
  value = aws_s3_bucket.state.id
}

output "lock_table" {
  value = aws_dynamodb_table.locks.name
}
