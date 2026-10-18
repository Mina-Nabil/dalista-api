# ------------------------------------------------------------------------------
# ALB security group — public HTTP/HTTPS inbound
# ------------------------------------------------------------------------------

resource "aws_security_group" "alb" {
  name        = "dalista-${var.environment}-alb"
  description = "Allow HTTPS/HTTP inbound to ALB"
  vpc_id      = aws_vpc.main.id

  tags = { Name = "dalista-${var.environment}-sg-alb" }
}

resource "aws_vpc_security_group_ingress_rule" "alb_https" {
  security_group_id = aws_security_group.alb.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 443
  to_port           = 443
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_ingress_rule" "alb_http" {
  security_group_id = aws_security_group.alb.id
  cidr_ipv4         = "0.0.0.0/0"
  from_port         = 80
  to_port           = 80
  ip_protocol       = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "alb_all" {
  security_group_id = aws_security_group.alb.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

# ------------------------------------------------------------------------------
# ECS tasks security group — inbound only from the ALB on 8080
# ------------------------------------------------------------------------------

resource "aws_security_group" "ecs" {
  name        = "dalista-${var.environment}-ecs"
  description = "Allow inbound from ALB on port 8080"
  vpc_id      = aws_vpc.main.id

  tags = { Name = "dalista-${var.environment}-sg-ecs" }
}

resource "aws_vpc_security_group_ingress_rule" "ecs_from_alb" {
  security_group_id            = aws_security_group.ecs.id
  referenced_security_group_id = aws_security_group.alb.id
  from_port                    = 8080
  to_port                      = 8080
  ip_protocol                  = "tcp"
}

resource "aws_vpc_security_group_egress_rule" "ecs_all" {
  security_group_id = aws_security_group.ecs.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

# ------------------------------------------------------------------------------
# RDS security group — inbound only from ECS tasks on 5432
# ------------------------------------------------------------------------------

resource "aws_security_group" "rds" {
  name        = "dalista-${var.environment}-rds"
  description = "Allow inbound from ECS on port 5432"
  vpc_id      = aws_vpc.main.id

  tags = { Name = "dalista-${var.environment}-sg-rds" }
}

resource "aws_vpc_security_group_ingress_rule" "rds_from_ecs" {
  security_group_id            = aws_security_group.rds.id
  referenced_security_group_id = aws_security_group.ecs.id
  from_port                    = 5432
  to_port                      = 5432
  ip_protocol                  = "tcp"
}
