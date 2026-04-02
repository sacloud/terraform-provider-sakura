data "sakura_apprun_dedicated_cluster" "main" {
  name = "ExampleCluster"
}

data "sakura_apprun_dedicated_application" "main" {
  id         = "A5F8D577-7395-4EB4-83D9-AC60A1EF2C5B"
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
}

resource "sakura_apprun_dedicated_version" "main" {
  application_id = data.sakura_apprun_dedicated_application.main.id
  cpu            = 1000
  memory         = 512
  image          = "nginx:latest"
  cmd            = ["/bin/sh"]
  scaling_mode   = "manual"
  fixed_scale    = 1
}
