terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }

  # Remote state in S3 with DynamoDB locking. The bucket and lock table are
  # created by ./bootstrap (run once, before `terraform init` here). State is
  # workspace-keyed:
  #   default    → staging (lands at staging/terraform.tfstate)
  #   production → env:/production/staging/terraform.tfstate
  backend "s3" {
    bucket         = "dalista-terraform-state"
    key            = "staging/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "dalista-terraform-locks"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "dalista"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

data "aws_caller_identity" "current" {}
