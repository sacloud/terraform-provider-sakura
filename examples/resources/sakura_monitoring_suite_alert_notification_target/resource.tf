resource "sakura_monitoring_suite_alert_notification_target" "foobar" {
  alert_project_id = "alert-project-resource-id" # e.g. sakura_monitoring_suite_alert_project.foobar.id
  description = "description"
  service_type = "simple_notification" # or "eventbus"
}