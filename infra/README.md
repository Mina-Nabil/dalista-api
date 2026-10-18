# Dalista API — Infrastructure (Terraform)

AWS-native, staging-only for Phase 0. Mirrors the Lumina reference layout, minus
its HIPAA/compliance stack (pgaudit, Vanta, GuardDuty, WAF, CloudTrail, VPC flow
logs) — those are deferred to Phase 12 hardening if needed.

## What this provisions

| File | Resources | Ticket |
|------|-----------|--------|
| `vpc.tf`, `security_groups.tf` | VPC, 2×public + 2×private subnets, IGW, NAT, route tables, SGs | DAL-5 |
| `rds.tf` | RDS Postgres 16 (private, encrypted, managed master password) | DAL-6 |
| `ecr.tf`, `ecs.tf`, `cloudwatch.tf` | ECR repo + lifecycle, ECS Fargate cluster, API service, migrate task, log group | DAL-7 |
| `alb.tf`, `acm.tf`, `route53.tf` | ALB + listeners, ACM cert (DNS-validated), Route53 A-record | DAL-8 |
| `secrets.tf` | Secrets Manager app secret (JWT, DB creds, integration placeholders) | DAL-9 |
| `iam.tf` | ECS execution/task roles, GitHub Actions OIDC provider + deploy role | DAL-7/CD |
| `bootstrap/` | S3 state bucket + DynamoDB lock table (local state) | DAL-10 |

## First-time setup

```bash
# 1. Create the remote-state backing store (runs once, uses local state).
cd infra/bootstrap
terraform init && terraform apply

# 2. Create the Route 53 hosted zone for your domain (manual) and note its ID.

# 3. Configure and apply the main staging infra.
cd ..
cp staging.tfvars.example staging.tfvars   # fill in domain, zone id, github ids
terraform init                              # uses the S3 backend from step 1
terraform apply -var-file=staging.tfvars
```

The `default` Terraform workspace is staging. Production (Phase 12) is a new
workspace: `terraform workspace new production` then apply with a
`production.tfvars`. Account-global resources (ECR repo, OIDC provider) are
created only in the `default` workspace and referenced elsewhere.

## Deploy wiring

`terraform output github_actions_role_arn` → set as the `AWS_DEPLOY_ROLE_ARN`
GitHub secret, consumed by `.github/workflows/deploy.yml`. A push to `master`
builds the image, pushes to ECR, runs the `dalista-staging-migrate` task, then
updates the `dalista-staging-api` service.

## Not applied here

No cloud credentials are configured in this environment, so this config has been
`terraform validate`-d and `fmt`-checked but **not** `apply`-d. Run the setup
steps above against a real AWS account to provision.
