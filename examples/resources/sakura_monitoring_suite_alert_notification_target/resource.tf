resource "sakura_monitoring_suite_alert_notification_target" "foobar" {
  alert_id = "alert-project-resource-id" # e.g. sakura_monitoring_suite_alert.foobar.id
  description = "description"
  service_type = "simple_notification" # or "eventbus"
}