resource "sakura_security_control_automated_action" "foobar" {
  name        = "foobar"
  description = "description"
  enabled     = true
  action = {
    type = "simple_notification"
    parameters = {
      service_principal_id = "your-service-principal-id",
      target_id = "your-simple-notification-group-id"
    }
    /* for workflows
    type = "workflows"
    parameters = {
      service_principal_id = "your-service-principal-id",
      target_id = "your-workflows-id"
      revision = 2
      args = "{\"sample\":10}"
      name = "from-tf"
    }
     */
  }
  execution_condition = "event.evaluationResult.status == \"REJECTED\""
}