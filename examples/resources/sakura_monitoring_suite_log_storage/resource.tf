resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "foobar"
  description = "description"
  classification = "shared" # or "dedicated"
  is_system = false
  # retention_period_days = 30 # default: 40
}