resource "sakura_vswitch" "foobar" {
	name = "foobar" 
	zone = "tk1b" # zone name # e.g. is1a 
}

resource "sakura_seg" "foobar" {
	vswitch_id  = sakura_vswitch.foobar.id
	zone        = sakura_vswitch.foobar.zone 
	server_ip_addresses = ["server_ip_address"] # e,g. 192.168.1.1 
	netmask     = 28 # 8-29
	endpoint_setting = {
		object_storage_endpoints = ["sakura-object-storage-endpoint"] # e.g. s3.isk01.sakurastorage.jp, s3.tky01.sakurastorage.jp
		monitoring_suite_endpoints = ["sakura-monitoring-suite-endpoint"] # e.g *****.logs.monitoring.global.api.salocloud.jp
		container_registry_endpoints = [""] # e.g. *****.sakuracr.jp
		ai_engine_endpoints = [""] # e.g. api.ai.sakura.ad.jp
		app_run_dedicated_control_enabled = true 
	}
	monitoring_suite_enabled = true
	dns_forwarding = {
		enabled = true
		private_hosted_zone = "example.com" # zone name. see sakura_dns/resource.tf
		upstream_dns_1 = "" # DNS Server (1) ns*.*****.sakura.ad.jp
		upstream_dns_2 = "" # DNS Server (2) ns*.*****.sakura.ad.jp
	}
}