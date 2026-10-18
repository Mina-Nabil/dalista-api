# ------------------------------------------------------------------------------
# DNS A-record (alias) → ALB
#
# The hosted zone is created manually in AWS and its ID passed via
# var.route53_zone_id.
# ------------------------------------------------------------------------------

resource "aws_route53_record" "api" {
  zone_id = var.route53_zone_id
  name    = "${var.api_subdomain}.${var.domain_name}"
  type    = "A"

  alias {
    name                   = aws_lb.main.dns_name
    zone_id                = aws_lb.main.zone_id
    evaluate_target_health = true
  }
}
