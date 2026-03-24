data "sakura_apprun_load_balancer_service_class" "example" {
  name = "example-class-name"
}

output "lb_service_class_path" {
  value = data.sakura_apprun_load_balancer_service_class.example.path
}

output "lb_service_class_node_count" {
  value = data.sakura_apprun_load_balancer_service_class.example.node_count
}