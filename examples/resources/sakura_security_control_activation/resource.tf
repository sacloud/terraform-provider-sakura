resource "sakura_security_control_activation" "foobar" {
  service_principal_id = "your-service-principal-id"
  enabled = true
  no_action_on_delete = true
}
