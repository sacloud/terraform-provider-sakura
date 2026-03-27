resource "sakura_monitoring_suite_log_routing" "foobar" {
  resource_id    = "target-resource-id" # e.g. sakura_simple_mq.foobar.id
  storage_id     = "log-storage-id"     # e.g. sakura_monitoring_suite_log_storage.foobar.id
  publisher_code = "service-name"       # e.g. "apprun", "database", "elb", etc...
  variant        = "log-variant"        # e.g. "applicationlog", "systemlog", etc, depends on publisher_code
}
