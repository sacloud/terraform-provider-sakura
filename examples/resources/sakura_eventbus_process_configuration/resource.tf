resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "foobar"
  description = "description"
  tags        = ["tag1"]

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"
  simplenotification_access_token_wo        = "test-token"
  simplenotification_access_token_secret_wo = "test-token-secret"
  credentials_wo_version                    = 1
  # or
  # destination = "simplemq"
  # parameters  = "{\"queue_name\": \"test-queue\", \"content\":\"TestContent\"}"
  # simplemq_api_key_wo    = "test-apikey"
}