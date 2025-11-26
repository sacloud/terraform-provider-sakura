data "sakura_cloudhsm" "foobar" {
  name = "foobar"
}

resource "sakura_cloudhsm_client" "foobar" {
  name = "foobar"
  cloudhsm_id = data.sakura_cloudhsm.foobar.id
  certificate = file("./client-cert.pem")
}
