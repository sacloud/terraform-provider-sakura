resource "sakura_object_storage_permission" "foobar" {
  name = "foobar"
  site_id = sakura_object_storage_bucket.foobar.site_id
  bucket_controls = [{
    bucket = sakura_object_storage_bucket.foobar.name
    can_read = true
    can_write = true
  }]
}