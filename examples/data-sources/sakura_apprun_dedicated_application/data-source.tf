data "sakura_apprun_dedicated_cluster" "main" {
  name = "Gkii8dvRskKjYOGzxL3D"
}

data "sakura_apprun_dedicated_application" "by_id" {
  id         = "A5F8D577-7395-4EB4-83D9-AC60A1EF2C5B"
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
}

data "sakura_apprun_dedicated_application" "by_name" {
  name       = "OghqEsr1JMB06MNKz3dc"
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
}
