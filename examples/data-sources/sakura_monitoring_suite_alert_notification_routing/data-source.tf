data "sakura_monitoring_suite_alert_notification_routing" "foobar" {
  id = "alert-notification-routing-uuid-id"
  alert_project_id = "alert-project-resource-id" # e.g. sakura_monitoring_suite_alert_project.foobar.id
}