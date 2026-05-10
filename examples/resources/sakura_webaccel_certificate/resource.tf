resource "sakura_webaccel_certificate" "foobar" {
  site_id             = "webaccel-site-id" # e.g. sakura_webaccel.foobar.id
  certificate_chain   = file("/path/to/crt.pem")
  private_key         = file("/path/to/key.pem")
  certificate_version = 1
}