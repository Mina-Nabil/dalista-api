# ------------------------------------------------------------------------------
# RDS subnet group (private subnets only)
# ------------------------------------------------------------------------------

resource "aws_db_subnet_group" "main" {
  name       = "dalista-${var.environment}"
  subnet_ids = [aws_subnet.private_a.id, aws_subnet.private_b.id]

  tags = { Name = "dalista-${var.environment}-db-subnet-group" }
}

# ------------------------------------------------------------------------------
# RDS PostgreSQL instance
#
# master password is auto-generated and stored in Secrets Manager by RDS
# (manage_master_user_password), then read back in secrets.tf to assemble the
# app's DB connection secrets. Multi-AZ only in the production workspace.
# ------------------------------------------------------------------------------

resource "aws_db_instance" "main" {
  identifier = "dalista-${var.environment}"

  engine         = "postgres"
  engine_version = "16"
  instance_class = "db.t4g.micro"

  allocated_storage = 20
  storage_type      = "gp3"
  storage_encrypted = true

  db_name  = "dalista_${var.environment}"
  username = "dalista_admin"

  manage_master_user_password = true

  enabled_cloudwatch_logs_exports = ["postgresql"]

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false
  multi_az               = terraform.workspace == "production"

  backup_retention_period = 7
  skip_final_snapshot     = terraform.workspace != "production"
  deletion_protection     = terraform.workspace == "production"

  tags = { Name = "dalista-${var.environment}-postgres" }
}
