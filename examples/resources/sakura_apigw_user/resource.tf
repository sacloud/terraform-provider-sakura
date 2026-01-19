resource "sakura_apigw_user" "foobar" {
  name = "foobar"
  tags = ["tag1"]
  custom_id = "custom-id-9999"
  ip_restriction = {
    protocols = "http"
    restricted_by = "allowIps"
    ips = ["192.168.0.10"]
  }
  groups = [{name = "test"}]
  authentication = {
    basic_auth = {
       username = "username",
       password_wo = "password"
       password_wo_version = 1
    },
    jwt = {
      key = "key",
      secret_wo = "secret",
      secret_wo_version = 1,
      algorithm = "HS256"
    },
    // hmac_auth can be set here as well
  }
}
