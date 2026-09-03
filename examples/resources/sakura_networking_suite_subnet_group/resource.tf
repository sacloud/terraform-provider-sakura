resource "sakura_networking_suite_subnet_group" "foobar" {
  name                    = "foobar"
  description             = "description"
  ipv4_address_range = "10.0.0.0/20"
  region                  = "is1"
}