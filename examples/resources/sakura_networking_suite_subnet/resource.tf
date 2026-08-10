resource "sakura_networking_suite_subnet" "foobar" {
  name                    = "foobar"
  description             = "description"
  ipv4_address_range_cidr = "10.0.0.0/24"
  zone                    = "is1c"
  subnet_group_srn        = sakura_networking_suite_subnet_group.foobar.srn
}