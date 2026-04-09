data "sakura_apprun_dedicated_cluster" "main" {
  name = "ExampleCluster"
}

resource "sakura_apprun_dedicated_application" "main" {
  cluster_id = sakura_apprun_dedicated_cluster.main.id
  name       = "ExampleApplication"
}
