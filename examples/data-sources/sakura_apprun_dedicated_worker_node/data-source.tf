data "sakura_apprun_dedicated_cluster" "main" {
  name = "Gkii8dvRskKjYOGzxL3D"
}

data "sakura_apprun_dedicated_auto_scaling_group" "main" {
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
  id         = "A5F8D577-7395-4EB4-83D9-AC60A1EF2C5B"
}

data "sakura_apprun_dedicated_worker_node" "main" {
  cluster_id            = data.sakura_apprun_dedicated_cluster.main.id
  auto_scaling_group_id = data.sakura_apprun_dedicated_auto_scaling_group.main.id
  id                    = "284BF902-A3D6-431F-B149-817B0971C9F9"
}
