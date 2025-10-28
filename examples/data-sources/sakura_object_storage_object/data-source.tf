data "sakura_object_storage_object" "foobar" {
  bucket  = data.sakura_object_storage_bucket.foobar.id
  key    = "myobject"
  access_key = sakura_object_storage_permission.foobar.access_key
  secret_key = sakura_object_storage_permission.foobar.secret_key
}