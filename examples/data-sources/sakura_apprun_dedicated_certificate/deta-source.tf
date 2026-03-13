data "sakura_apprun_dedicated_cluster" "main" {
  name = "Gkii8dvRskKjYOGzxL3D"
}

data "sakura_apprun_dedicated_certificate" "by_id" {
  id         = "B46841E8-73EB-483E-AF0A-8C7BB0EA63B6"
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
}

data "sakura_apprun_dedicated_certificate" "by_name" {
  name       = "KuhnzDtcMtsU9Pa5qsnF"
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
}
