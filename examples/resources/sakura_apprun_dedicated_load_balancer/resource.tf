locals {
  sakura_dns = [ "133.242.0.3", "133.242.0.4" ]
}

data "sakura_zone" "is1c" {
  name = "is1c"
}

data "sakura_apprun_dedicated_cluster" "main" {
  name = "Gkii8dvRskKjYOGzxL3D"
}

data "sakura_apprun_dedicated_auto_scaling_group" "main" {
  cluster_id = data.sakura_apprun_dedicated_cluster.main.id
  name       = "HejIrLkM2DWO8UPQvGOw"
}

data "sakura_apprun_dedicated_load_balancer_service_classes" "main" {}

data "sakura_internet" "main" {
  name = "ozoQuqtyLG022lA3C1Nv"
}

resource "sakura_apprun_dedicated_load_balancer" "main" {
  cluster_id            = data.sakura_apprun_dedicated_cluster.main.id
  auto_scaling_group_id = data.sakura_apprun_dedicated_auto_scaling_group.main.id
  name                  = "Pg065g2wSVJa3DSWxCDO"
  service_class_path    = data.sakura_apprun_dedicated_load_balancer_service_classes.main.classes[0].path
  name_servers          = local.sakura_dns

  interfaces = [{
    interface_index   = 0
    upstream          = data.sakura_internet.main.vswitch_id
    ip_pool = [{
      start = data.sakura_internet.main.min_ip_address
      end   = data.sakura_internet.main.max_ip_address
    }]
    netmask_len       = data.sakura_internet.main.netmask
    default_gateway   = data.sakura_internet.main.gateway
    vip               = cidrhost("${data.sakura_internet.main.gateway}/${data.sakura_internet.main.netmask}", 9)
    virtual_router_id = 1
  }]
}
