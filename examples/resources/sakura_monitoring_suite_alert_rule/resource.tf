resource "sakura_monitoring_suite_alert_rule" "foobar" {
  name = "foobar"
  alert_project_id = "alert-project-resource-id" # e.g. sakura_monitoring_suite_alert_project.foobar.id
  metric_storage_id = "metric-storage-resource-id" # e.g. sakura_monitoring_suite_metric_storage.foobar.id 
  query = "count_values" # If your query includes a newline at the end, e.g. use here document or file, call chomp function to remove a newline.
  enabled_warning = true
  enabled_critical = true
  threshold_warning = ">= 10"
  threshold_critical = ">= 20"
  threshold_duration_warning = 600
  threshold_duration_critical = 600
}