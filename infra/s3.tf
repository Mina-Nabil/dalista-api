# ------------------------------------------------------------------------------
# S3 bucket for user-uploaded item images (storage driver = s3).
#
# Objects are public-read (served directly as the item image URL). Bucket name
# is suffixed with the account id for global uniqueness. Per-environment bucket.
# ------------------------------------------------------------------------------

resource "aws_s3_bucket" "media" {
  bucket = "dalista-${var.environment}-media-${data.aws_caller_identity.current.account_id}"

  tags = { Name = "dalista-${var.environment}-media" }
}

# Allow a public bucket policy (but keep ACLs disabled — access is granted by the
# policy below, not per-object ACLs).
resource "aws_s3_bucket_public_access_block" "media" {
  bucket = aws_s3_bucket.media.id

  block_public_acls       = true
  ignore_public_acls      = true
  block_public_policy     = false
  restrict_public_buckets = false
}

# Public read of objects so returned image URLs resolve without signing.
resource "aws_s3_bucket_policy" "media_public_read" {
  bucket = aws_s3_bucket.media.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "PublicReadGetObject"
      Effect    = "Allow"
      Principal = "*"
      Action    = "s3:GetObject"
      Resource  = "${aws_s3_bucket.media.arn}/*"
    }]
  })

  depends_on = [aws_s3_bucket_public_access_block.media]
}

# Browser clients (dashboard) fetch images cross-origin; allow GET.
resource "aws_s3_bucket_cors_configuration" "media" {
  bucket = aws_s3_bucket.media.id

  cors_rule {
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = ["*"]
    allowed_headers = ["*"]
    max_age_seconds = 3600
  }
}
