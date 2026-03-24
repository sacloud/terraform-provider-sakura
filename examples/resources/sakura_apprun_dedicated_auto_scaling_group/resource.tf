locals {
  sakura_dns = [ "133.242.0.3", "133.242.0.4" ]
}

data "sakura_zone" "is1c" {
  name = "is1c"
}

data "sakura_internet" "main" {
  name = "ozoQuqtyLG022lA3C1Nv"
}

data "sakura_apprun_dedicated_cluster" "main" {
  name = "Gkii8dvRskKjYOGzxL3D"
}

data "sakura_apprun_dedicated_worker_service_classes" "main" {}

resource "sakura_apprun_dedicated_auto_scaling_group" "main" {
  cluster_id                = sakura_apprun_dedicated_cluster.main.id
  name                      = "Pg065g2wSVJa3DSWxCDO"
  zone                      = data.sakura_zone.is1c.name
  worker_service_class_path = data.sakura_apprun_dedicated_worker_service_classes.main.classes[0].path
  name_servers              = local.sakura_dns
  min_nodes                 = 1
  max_nodes                 = 3

  interfaces = [{
    interface_index = 0
    upstream        = data.sakura_internet.main.vswitch_id
    connects_to_lb  = false
    netmask_len     = data.sakura_internet.main.netmask
    default_gateway = data.sakura_internet.main.gateway
    ip_pool = [{
      start = data.sakura_internet.main.min_ip_address
      end   = data.sakura_internet.main.max_ip_address
    }]
  }]
}
