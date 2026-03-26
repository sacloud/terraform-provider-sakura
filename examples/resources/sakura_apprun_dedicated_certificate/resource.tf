data "sakura_apprun_dedicated_cluster" "main" {
  name = "zJFhR15eLd17ZmeOK0Xw"
}

resource "sakura_apprun_dedicated_certificate" "main" {
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
  name       = "KuhnzDtcMtsU9Pa5qsnF"

  certificate_pem              = file("${path.module}/cert.pem")
  private_key_pem              = file("${path.module}/key.pem")
  intermediate_certificate_pem = file("${path.module}/intermediate.pem")
}
