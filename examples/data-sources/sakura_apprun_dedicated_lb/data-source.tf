data "sakura_apprun_dedicated_cluster" "main" {
  name = "ExampleCluster"
}

data "sakura_apprun_dedicated_auto_scaling_group" "main" {
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
  name       = "ExampleASG"
}

data "sakura_apprun_dedicated_lb" "by_id" {
  cluster_id            = data.sakura_apprun_dedicated_cluster.main.id
  auto_scaling_group_id = data.sakura_apprun_dedicated_auto_scaling_group.main.id
  id                    = "A5F8D577-7395-4EB4-83D9-AC60A1EF2C5B"
}

data "sakura_apprun_dedicated_lb" "by_name" {
  cluster_id            = data.sakura_apprun_dedicated_cluster.main.id
  auto_scaling_group_id = data.sakura_apprun_dedicated_auto_scaling_group.main.id
  name                  = "ExampleLB"
}
