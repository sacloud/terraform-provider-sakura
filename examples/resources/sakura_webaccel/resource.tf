resource sakura_webaccel "foobar" {
  name = "foobar"
  domain_type = "subdomain" # or "own_domain" with domain attribute
  request_protocol = "https"
  // origin configuration
  origin_parameters = {
    type = "web" # or "bucket" with Object Storage attributes
    origin = "foobar.example.com"
    host_header = "foobar.example.com"
    protocol = "https"
  }
  // logging configuration
  logging = {
    enabled = true
    endpoint = "s3.isk01.sakurastorage.jp"
    region = "jp-north-1"
    bucket_name = "foobar-bucket"
    access_key = "..."
    secret_access_key = "..."
    credentials_version = 1
  }
  // other parameters, e.g. vary_support, cors_rules, etc.
}