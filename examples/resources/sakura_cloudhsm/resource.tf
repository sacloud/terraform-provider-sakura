resource "sakura_cloudhsm" "foobar" {
  name = "foobar"
  description = "foobar description"
  tags = ["terraform"]
  ipv4_network_address = "192.168.50.0"
  ipv4_netmask = 28
}
