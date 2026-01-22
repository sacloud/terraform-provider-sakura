resource "sakura_apigw_cert" "foobar" {
  name = "foobar"
  rsa  = {
    cert_wo = file("path/to/rsa.crt")
    key_wo = file("path/to/rsa.key")
    cert_wo_version = 1
  }
  ecdsa = {
    cert_wo = file("path/to/ecdsa.crt")
    key_wo = file("path/to/ecdsa.key")
    cert_wo_version = 1
  }
}
