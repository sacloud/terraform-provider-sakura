resource "sakura_networking_suite_interface_connection" "foobar" {
  subnet_srn             = sakura_networking_suite_subnet.foobar.srn
  interface_srn          = "<your-server-interface-srn>" # e.g. "srnv1:sakura-is1c:sakura.iaas.interface:123456789012"
  ephemeral_ipv4_address = "10.0.0.5" # if you omit this attribute, the networking-suite assigns one automatically
}