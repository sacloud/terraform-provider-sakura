data "sakura_apigw_subscription" "foobar" {
  name = "test"
}

resource "sakura_apigw_service" "foobar" {
  name     = "foobar"
  tags     = ["tag1", "tag2"]
  protocol = "https"
  host     = "foobar.example.com"
  subscription_id = data.sakura_apigw_subscription.foobar.id
  /* CORS settings
  cors_config = {
    access_control_allow_methods = ["GET", "POST", "OPTIONS"]
    access_control_allow_headers = "*"
    max_age = 3600
  }
  */
  /* Integration with object storage
  object_storage_config = {
    bucket = "test"
    region = data.sakura_object_storage_site.foobar.region
    endpoint = data.sakura_object_storage_site.foobar.s3_endpoint
    access_key_wo = "your-access-key"
    secret_access_key_wo = "your-secret-access-key"
    credentials_wo_version = 1
  }
  */
}