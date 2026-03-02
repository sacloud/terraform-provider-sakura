resource "sakura_addon_search" "foobar" {
  location = "japaneast"
  partition_count = 1
  replica_count = 1
  sku = 2
}