resource "sakura_monitoring_suite_log_measure_rule" "foobar" {
  name = "foobar"
  description = "description"
  alert_id = "alert-project-resource-id" # e.g. sakura_monitoring_suite_alert.foobar.id
  log_storage_id = "log-storage-resource-id" # e.g. sakura_monitoring_suite_log_storage.foobar.id
  metric_storage_id = "metric-storage-resource-id" # e.g. sakura_monitoring_suite_metric_storage.foobar.id 
  rule = {
    version = "v1"
    query = {
      # See https://manual.sakura.ad.jp/api/cloud/portal/?api=monitoring-suite-api#tag/%E3%82%A2%E3%83%A9%E3%83%BC%E3%83%88/operation/alerts_projects_log_measure_rules_create for details of the query matchers.
      matchers = jsonencode([
        {
          type = "string"
          field = "text_payload"
          value = "value"
          operator = "eq"
          value_list = []
        }
      ])
      /*
       * Complex matchers can be defined by nesting matchers.
      matchers = jsonencode([
        {
          type = "or" # or "and"
          matchers = [
            {
              type = "string"
              # ...
            },
            {
              type = "number"
              # ...
            }
          ]
        }
      ])
      */
    }
  }
}