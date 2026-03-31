data "sakura_iam_service_principal" "main" {
  name = "ExampleSP"
}

resource "sakura_apprun_dedicated_cluster" "main" {
  name                 = "ExampleCluster"
  service_principal_id = data.sakura_iam_service_principal.main.id
  lets_encrypt_email   = "tf-test-email@sakura.ad.jp"

  ports = [
    {
      port     = 80
      protocol = "http"
    },
  ]
}
