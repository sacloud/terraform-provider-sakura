resource "sakura_webaccel_certificate" "foobar" {
  site_id                = "webaccel-site-id" # e.g. sakura_webaccel.foobar.id
  certificate_chain_wo   = file("/path/to/crt.pem")
  private_key_wo         = file("/path/to/key.pem")
  certificate_wo_version = 1
}