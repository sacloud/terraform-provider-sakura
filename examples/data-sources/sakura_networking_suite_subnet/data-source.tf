data "sakura_networking_suite_subnet" "foobar" {
  srn = "subnet-srn" # e.g. srnv1:sakura-is1c:sakura.networking-suite.subnet:2345678901
  # or name based search. For name based search, you must specify the subnet_group_srn together.
  # name = "foobar"
  # subnet_group_srn = "subnet-group-srn" # e.g. srnv1:sakura-is1:sakura.networking-suite.subnet-group:1234567890
}