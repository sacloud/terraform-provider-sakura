data "sakura_apprun_dedicated_cluster" "main" {
  name = "zJFhR15eLd17ZmeOK0Xw"
}

resource "sakura_apprun_dedicated_application" "main" {
  cluster_id = sakura_apprun_dedicated_cluster.main.id
  name       = "jyrrYrG6ptdNujmTV7CR"
}