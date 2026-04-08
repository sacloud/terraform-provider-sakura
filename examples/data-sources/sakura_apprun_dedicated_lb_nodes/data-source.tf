data "sakura_apprun_dedicated_cluster" "main" {
  name = "ExampleCluster"
}

data "sakura_apprun_dedicated_auto_scaling_group" "main" {
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
  id         = "A5F8D577-7395-4EB4-83D9-AC60A1EF2C5B"
}

data "sakura_apprun_dedicated_lb" "main" {
  cluster_id            = data.sakura_apprun_dedicated_cluster.main.id
  auto_scaling_group_id = data.sakura_apprun_dedicated_auto_scaling_group.main.id
  name                  = "ExampleLB"
}

data "sakura_apprun_dedicated_lb_nodes" "main" {
  cluster_id            = data.sakura_apprun_dedicated_cluster.main.id
  auto_scaling_group_id = data.sakura_apprun_dedicated_auto_scaling_group.main.id
  lb_id      = data.sakura_apprun_dedicated_lb.main.id
}
