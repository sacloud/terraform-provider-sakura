resource "sakura_dedicated_storage" "foobar" {
  name        = "foobar"
  description = "description"
  tags        = ["tag1", "tag2", "tag3"]
  # icon_id     = sakura_icon.foobar.id
}