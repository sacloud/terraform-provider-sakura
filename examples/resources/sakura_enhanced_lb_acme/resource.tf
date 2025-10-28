resource "sakura_enhanced_lb_acme" "foobar" {
  proxylb_id        = sakura_enhanced_lb.foobar.id
  accept_tos        = true
  common_name       = "www.example.com"
  subject_alt_names = ["www1.example.com"]

  update_delay_sec             = 120
  get_certificates_timeout_sec = 120
}

data "sakura_enhanced_lb" "foobar" {
    name = "foobar"
}