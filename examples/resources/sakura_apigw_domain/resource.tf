data "sakura_apigw_cert" "foobar" {
  name = "foobar"
}

resource "sakura_apigw_domain" "foobar" {
  name           = "foobar.example.com"
  certificate_id = data.sakura_apigw_cert.foobar.id
}