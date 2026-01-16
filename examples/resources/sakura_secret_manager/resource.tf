resource "sakura_secret_manager" "foobar" {
  name        = "foobar"
  description = "description"
  tags        = ["tag1", "tag2"]
  kms_key_id  = "kms-resource-id" # e.g. sakura_kms.foobar.id
}
