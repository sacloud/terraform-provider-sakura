data "sakura_cloudhsm" "foobar" {
  name = "foobar"
}

resource "sakura_cloudhsm_peer" "foobar" {
  cloudhsm_id = data.sakura_cloudhsm.foobar.id
  router_id  = "local-router-id"
  secret_key = "local-router-secret-key"
  # TODO: Use local_router resource/data source when supported
}
