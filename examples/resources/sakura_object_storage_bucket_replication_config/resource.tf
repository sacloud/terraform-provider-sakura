resource "sakura_object_storage_bucket_replication_config" "foobar" {
  site_id = "isk01"  # e.g. data.sakura_object_storage_site.isk.id
  bucket  = "foobar" # e.g. sakura_object_storage_bucket.foobar1.name
  destination_bucket = "dest-foobar" # e.g. sakura_object_storage_bucket.foobar2.name
}