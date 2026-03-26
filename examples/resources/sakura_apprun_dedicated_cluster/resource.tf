data "sakura_iam_service_principal" "main" {
  name = "m2TCFudC7isJ5fDAoKOeXq0Be3SD36m55OTh3NDhKvVdQZRm4hMROqY1TH8jDvP4"
}

resource "sakura_apprun_dedicated_cluster" "main" {
  name                 = "Gkii8dvRskKjYOGzxL3D"
  service_principal_id = data.sakura_iam_service_principal.main.id
  lets_encrypt_email   = "tf-test-email@sakura.ad.jp"

  ports = [
    {
      port     = 80
      protocol = http
    },
  ]
}
