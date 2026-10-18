variable "environment" {
  description = "Environment name (staging or production)"
  type        = string
  default     = "staging"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

# ------------------------------------------------------------------------------
# DNS
# ------------------------------------------------------------------------------

variable "domain_name" {
  description = "Root domain name (e.g. dalista.app)"
  type        = string
}

variable "api_subdomain" {
  description = "Subdomain for the API (e.g. 'api-staging' for staging, 'api' for production)"
  type        = string
  default     = "api-staging"
}

variable "route53_zone_id" {
  description = "Route 53 hosted zone ID for the domain (created manually; ID passed in)"
  type        = string
}

variable "front_end_url" {
  description = "Public dashboard base URL (no trailing slash), used for CORS. Passed to the API as FRONT_END_URL."
  type        = string
  default     = ""
}

# ------------------------------------------------------------------------------
# API runtime config
# ------------------------------------------------------------------------------

variable "jwt_expiry_hours" {
  description = "JWT token expiry in hours"
  type        = number
  default     = 8
}

variable "rate_limit_per_minute" {
  description = "Per-IP request limit per minute (token bucket, burst = limit). 0 disables."
  type        = number
  default     = 40
}

variable "sentry_dsn" {
  description = "Sentry DSN for error tracking"
  type        = string
  sensitive   = true
  default     = ""
}

# ------------------------------------------------------------------------------
# Integration credentials — stored in Secrets Manager, injected into ECS tasks.
# Optional for Phase 0 bring-up; populated as each feature phase lands. Never
# commit real values — pass via a gitignored *.tfvars.
# ------------------------------------------------------------------------------

variable "sms_misr_username" {
  description = "SMS Misr API username (OTP + admin SMS)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "sms_misr_password" {
  description = "SMS Misr API password"
  type        = string
  sensitive   = true
  default     = ""
}

variable "sms_misr_sender" {
  description = "SMS Misr sender ID"
  type        = string
  default     = ""
}

variable "fcm_credentials_json" {
  description = "Firebase Cloud Messaging service-account JSON (push notifications)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "apple_iap_private_key" {
  description = "Apple App Store Server API private key (.p8 contents)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "google_play_credentials_json" {
  description = "Google Play Developer API service-account JSON"
  type        = string
  sensitive   = true
  default     = ""
}

# ------------------------------------------------------------------------------
# GitHub Actions OIDC (deploy role trust)
# ------------------------------------------------------------------------------

variable "github_org" {
  description = "GitHub organization or username for OIDC trust"
  type        = string

  validation {
    condition     = !can(regex("/", var.github_org))
    error_message = "github_org must be a bare org slug, not a URL or path."
  }
}

variable "github_repo" {
  description = "GitHub repository name for OIDC trust"
  type        = string
}

variable "github_org_id" {
  description = "Numeric GitHub organization ID (gh api orgs/<org> --jq .id). Pins the OIDC subject claim across org renames."
  type        = string
}

variable "github_repo_id" {
  description = "Numeric GitHub repository ID (gh api repos/<org>/<repo> --jq .id)."
  type        = string
}
