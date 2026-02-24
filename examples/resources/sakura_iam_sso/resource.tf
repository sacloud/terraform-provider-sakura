resource "sakura_iam_sso" "foobar" {
  name = "foobar"
  description = "description"
  idp_entity_id = "https://idp.example.com/ile2ephei7saeph6"
  idp_login_url = "https://idp.example.com/ile2ephei7saeph6/sso/login"
  idp_logout_url = "https://idp.example.com/ile2ephei7saeph6/sso/logout"
  idp_certificate = file("rsa.crt")
}