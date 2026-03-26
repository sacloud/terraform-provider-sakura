data "sakura_apprun_dedicated_cluster" "main" {
  name = "Gkii8dvRskKjYOGzxL3D"
}

data "sakura_apprun_dedicated_auto_scaling_group" "main" {
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
  name       = "HejIrLkM2DWO8UPQvGOw"
}

data "sakura_apprun_dedicated_load_balancers" "main" {
  cluster_id            = data.sakura_apprun_dedicated_cluster.main.id
  auto_scaling_group_id = data.sakura_apprun_dedicated_auto_scaling_group.main.id
}
