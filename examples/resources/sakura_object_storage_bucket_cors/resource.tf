resource "sakura_object_storage_bucket_cors" "foobar" {
  bucket  = sakura_object_storage_bucket.foobar.name
  access_key = sakura_object_storage_permission.foobar.access_key
  secret_key = sakura_object_storage_permission.foobar.secret_key
  cors_rules = [{
    allowed_methods = ["GET", "PUT"]
    allowed_origins = ["https://obj-storage.example.jp"]
    expose_headers = ["ETag"]
  }]
}