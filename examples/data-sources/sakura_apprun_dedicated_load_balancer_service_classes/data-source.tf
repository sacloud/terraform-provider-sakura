data "sakura_apprun_load_balancer_service_classes" "main" {}

output "lb_service_class_names" {
  value = data.sakura_apprun_load_balancer_service_classes.main.classes[*].name
}

output "lb_service_class_paths" {
  value = data.sakura_apprun_load_balancer_service_classes.main.classes[*].path
}
