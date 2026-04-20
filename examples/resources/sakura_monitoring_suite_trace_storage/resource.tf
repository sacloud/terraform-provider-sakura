resource "sakura_monitoring_suite_trace_storage" "foobar" {
  name = "foobar"
  description = "description"
  # retention_period_days = 30 # default: 40
}
