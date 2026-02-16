resource sakura_server "server" {
  name = "foobar"
  network_interface = [{
    upstream = "shared"
  }]
}

resource "sakura_ipv4_ptr" "foobar" {
  ip_address     = sakura_server.server.ip_address
  hostname       = "www.example.com"
  retry_max      = 30
  retry_interval = 10
}