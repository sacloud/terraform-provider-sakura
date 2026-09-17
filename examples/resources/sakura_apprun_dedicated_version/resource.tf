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

  env_vars = [
    {
      key   = "LOG_LEVEL"
      value = "info"
    },
  ]

  secret_vars = [
    {
      key              = "API_TOKEN"
      value_wo         = "s3cr3t" # write-only: never stored in the state
      value_wo_version = 1        # bump this to create a new version with a new value_wo
    },
  ]
}
