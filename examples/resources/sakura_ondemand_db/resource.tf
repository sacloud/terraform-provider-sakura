resource "sakura_ondemand_db" "foobar" {
  name          = "example"
  database_name = "exampledb"
  database_type = "tidb"
  region        = "is1"  # or "tk1"
  password_wo   = "password-123456789"
  password_wo_version = 1

  description = "description"
  tags        = ["tag1", "tag2"]
}