data "sakura_apprun_load_balancer_service_class" "example" {
  name = "AppRun専有型 ロードバランサ 2vCPU / 2GBメモリ（冗長構成）"
}

output "lb_service_class_path" {
  value = data.sakura_apprun_load_balancer_service_class.example.path
}

output "lb_service_class_node_count" {
  value = data.sakura_apprun_load_balancer_service_class.example.node_count
}
