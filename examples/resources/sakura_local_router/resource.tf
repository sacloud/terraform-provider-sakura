data "sakura_vswitch" "foobar" {
  name = "foobar"
}

data "sakura_local_router" "peer" {
  name = "peer"
}

resource "sakura_local_router" "example" {
  name        = "example"
  description = "description"
  tags        = ["tag1", "tag2"]

  // Since it can connect to switch of services other than sakura_vswitch,
  // use the parameter name as "switch".
  switch = {
    code     = data.sakura_vswitch.foobar.id
    category = "cloud"
    zone     = "is1a"
  }

  network_interface = {
    vip          = "192.168.11.1"
    ip_addresses = ["192.168.11.11", "192.168.11.12"]
    netmask      = 24
    vrid         = 101
  }

  static_route = [{
    prefix   = "10.0.0.0/24"
    next_hop = "192.168.11.2"
  },
  {
    prefix   = "172.16.0.0/16"
    next_hop = "192.168.11.3"
  }]

  peer = [{
    peer_id     = data.sakura_local_router.peer.id
    secret_key  = data.sakura_local_router.peer.secret_keys[0]
    description = "description"
  }]
}
