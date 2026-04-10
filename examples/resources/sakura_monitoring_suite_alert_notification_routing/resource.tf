resource "sakura_monitoring_suite_alert_notification_routing" "foobar" {
  alert_id = "alert-project-resource-id" # e.g. sakura_monitoring_suite_alert.foobar.id
  notification_target_id = "notification-target-resource-id" # e.g. sakura_monitoring_suite_alert_notification_target.foobar.id
  resend_interval_minutes = 60
  match_labels = [{
    "name" = "foo"
    "value" = "bar"
  }]
}