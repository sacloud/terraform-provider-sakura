resource "sakura_private_host" "foobar" {
  name        = "foobar"
  description = "description"
  tags        = ["tag1", "tag2"]

  # dedicated_storage_id = sakura_dedicated_storage.foobar.id
}