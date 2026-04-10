data "sakura_monitoring_suite_alert_log_measure_rule" "foobar" {
  name = "foobar"
  # or
  # id = "alert-log-measure-rule-uuid-id"
  alert_id = "alert-project-resource-id" # e.g. sakura_monitoring_suite_alert.foobar.id
}