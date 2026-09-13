resource "sakura_vswitch" "foobar" {
	name = "foobar" 
	zone = "tk1b" # zone name # e.g. is1a 
}

resource "sakura_seg" "foobar" {
	vswitch_id  = sakura_vswitch.foobar.id
	zone        = sakura_vswitch.foobar.zone 
	server_ip_addresses = ["192.168.1.1"]
	netmask     = 28 # 8-29
	endpoint_setting = {
		object_storage_endpoints = ["s3.isk01.sakurastorage.jp"] # tky01 and arc02 are alsosupported
		monitoring_suite_endpoints = ["*****.logs.monitoring.global.api.salocloud.jp"] # metrics is also supported
		container_registry_endpoints = ["*****.sakuracr.jp"]
		ai_engine_endpoints = ["api.ai.sakura.ad.jp"]  # only api.ai.sakura.ad.jp is supported
		simple_ai_endpoints = ["simpleai.is1.api.sacloud.jp"] # only simpleai.is1.api.sacloud.jp is supported
		app_run_dedicated_control_enabled = true 
	}
	monitoring_suite_enabled = true
	dns_forwarding = {
		enabled = true
		private_hosted_zone = "example.com" # zone name. see sakura_dns/resource.tf
		dns_servers = ["ns*.*****.sakura.ad.jp","ns*.*****.sakura.ad.jp"] # Must set two record.
	}
}