resource "sakura_monitoring_suite_metric_routing" "foobar" {
  resource_id    = "target-resource-id" # e.g. sakura_simple_mq.foobar.id
  storage_id     = "metric-storage-id"  # e.g. sakura_monitoring_suite_metric_storage.foobar.id
  publisher_code = "service-name"       # e.g. "apprun", "database", "nfs", etc...
  variant        = "metric-variant"     # e.g. "applicationmetrics", "systemmetrics", etc, depends on publisher_code
}
