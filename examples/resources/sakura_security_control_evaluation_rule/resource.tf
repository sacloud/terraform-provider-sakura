// service_principal_id and zones/sites required resources
resource "sakura_security_control_evaluation_rule" "foobar1" {
  id         = "server-no-public-ip"
  enabled    = true
  parameters = {
    service_principal_id = "your-service-principal-id"
    targets = ["is1a", "tk1b"]  // [] for all zones
  }
  no_action_on_delete = true
}

// service_principal_id only required resources
resource "sakura_security_control_evaluation_rule" "foobar2" {
  id         = "elb-logging-enabled"
  enabled    = true
  parameters = {
    service_principal_id = "your-service-principal-id"
  }
  no_action_on_delete = true
}

// no parameters required resources
resource "sakura_security_control_evaluation_rule" "foobar3" {
  id      = "addon-threat-detections"
  enabled = true
  no_action_on_delete = true
}