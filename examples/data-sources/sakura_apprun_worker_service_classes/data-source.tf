data "sakura_apprun_worker_service_classes" "main" {}

output "worker_service_class_names" {
  value = data.sakura_apprun_worker_service_classes.main.classes[*].name
}

output "worker_service_class_paths" {
  value = data.sakura_apprun_worker_service_classes.main.classes[*].path
}