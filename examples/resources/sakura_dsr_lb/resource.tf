resource "sakura_vswitch" "foobar" {
  name = "foobar"
}

resource "sakura_dsr_lb" "foobar" {
  name        = "foobar"
  description = "description"
  tags        = ["tag1", "tag2"]

  plan = "standard"

  network_interface = {
    vswitch_id   = sakura_vswitch.foobar.id
    vrid         = 1
    ip_addresses = ["192.168.11.101"]
    netmask      = 24
    gateway      = "192.168.11.1"
  }

  vip = [{
    vip          = "192.168.11.201"
    port         = 80
    delay_loop   = 10
    sorry_server = "192.168.11.21"

    server = [{
      ip_address = "192.168.11.51"
      protocol   = "http"
      path       = "/health"
      status     = 200
    },{
      ip_address = "192.168.11.52"
      protocol   = "http"
      path       = "/health"
      status     = 200
    }]
  }]
}
