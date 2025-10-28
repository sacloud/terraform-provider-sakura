resource "sakura_object_storage_bucket" "foobar" {
  name    = "foobar"
  site_id = data.sakura_object_storage_site.foobar.id
}