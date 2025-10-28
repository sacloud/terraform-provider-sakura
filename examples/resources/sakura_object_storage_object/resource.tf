resource "sakura_object_storage_object" "foobar" {
  bucket = sakura_object_storage_bucket.foobar.name
  key    = "foobar.txt"
  access_key = sakura_object_storage_permission.foobar.access_key
  secret_key = sakura_object_storage_permission.foobar.secret_key
  content = "Hello World!"
}