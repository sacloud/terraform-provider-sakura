resource "sakura_object_storage_bucket_encryption_config" "foobar" {
  site_id     = "tky01"  # e.g. data.sakura_object_storage_site.tky.id
  bucket      = "foobar" # e.g. sakura_object_storage_bucket.foobar.name
  kms_key_id  = "kms-resource-id" # e.g. sakura_kms.foobar.id
}