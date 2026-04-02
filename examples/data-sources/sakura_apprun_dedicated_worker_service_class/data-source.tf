data "sakura_apprun_worker_service_class" "example" {
  name = "AppRun専有型 ワーカ 8vCPU / 8GBメモリ"
}

output "worker_service_class_path" {
  value = data.sakura_apprun_worker_service_class.example.path
}
