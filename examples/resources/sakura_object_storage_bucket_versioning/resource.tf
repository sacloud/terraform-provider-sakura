resource "sakura_object_storage_bucket_versioning" "foobar" {
  bucket  = sakura_object_storage_bucket.foobar.name
  access_key = sakura_object_storage_permission.foobar.access_key
  secret_key = sakura_object_storage_permission.foobar.secret_key
  versioning_configuration = {
    status = "Enabled"
  }
}
