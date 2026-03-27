resource "sakura_monitoring_suite_log_storage_access_key" "foobar" {
  storage_id = "log-storage-resource-id" # e.g. sakura_monitoring_suite_log_storage.foobar.id
  description = "description"
}
