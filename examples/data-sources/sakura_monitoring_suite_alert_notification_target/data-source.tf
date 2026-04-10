data "sakura_monitoring_suite_alert_notification_target" "foobar" {
  id = "alert-notification-target-uuid-id"
  alert_id = "alert-project-resource-id" # e.g. sakura_monitoring_suite_alert.foobar.id
}