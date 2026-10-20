data "sakura_apprun_dedicated_cluster" "main" {
  name = "ExampleCluster"
}

resource "sakura_apprun_dedicated_application" "main" {
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
  name       = "ExampleApplication"
}

resource "sakura_apprun_dedicated_active_version" "main" {
  application_id = sakura_apprun_dedicated_application.main.id
  cpu            = 1000
  memory         = 1024
  scaling_mode   = "manual"
  fixed_scale    = 1
  image          = "nginx:latest"

  exposed_ports = [
    {
      target_port = 80
    }
  ]
}
