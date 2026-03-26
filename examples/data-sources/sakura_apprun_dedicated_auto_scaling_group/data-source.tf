data "sakura_apprun_dedicated_cluster" "main" {
  name = "Gkii8dvRskKjYOGzxL3D"
}

data "sakura_apprun_dedicated_auto_scaling_group" "by_id" {
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
  id         = "A5F8D577-7395-4EB4-83D9-AC60A1EF2C5B"
}

data "sakura_apprun_dedicated_auto_scaling_group" "by_name" {
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
  name       = "HejIrLkM2DWO8UPQvGOw"
}
