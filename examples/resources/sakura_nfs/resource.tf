resource "sakura_nfs" "foobar" {
  name = "foobar"
  plan = "ssd"
  size = "500"

  network_interface = {
    switch_id   = sakura_switch.foobar.id
    ip_address  = "192.168.11.101"
    netmask     = 24
    gateway     = "192.168.11.1"
  }

  description = "description"
  tags        = ["tag1", "tag2"]
}

resource "sakura_switch" "foobar" {
  name = "foobar"
}