data "sakura_apprun_worker_service_class" "example" {
  name = "example-class-name"
}

output "worker_service_class_path" {
  value = data.sakura_apprun_worker_service_class.example.path
}
