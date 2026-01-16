resource "sakura_secret_manager_secret" "foobar" {
  name     = "foobar"
  vault_id = "secret_manager-resource-id" # e.g. sakura_secret_manager.foobar.id
  value_wo = "secret value!"
  value_wo_version = 1
  // for backward compatibility
  //value = "secret value!"
}