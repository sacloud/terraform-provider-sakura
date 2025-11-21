data "sakura_cloudhsm" "foobar" {
  name = "foobar"
}

data "sakura_cloudhsm_peer" "foobar" {
  id = "peer-resource-id" # This ID is same as local-router ID used in peer
  cloudhsm_id = data.sakura_cloudhsm.foobar.id
}
