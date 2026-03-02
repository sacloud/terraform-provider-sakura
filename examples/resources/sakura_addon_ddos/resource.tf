resource "sakura_addon_ddos" "foobar" {
  location = "japaneast"
  pricing_level = 1
  patterns = ["/*"]
  origin = {
    hostname = "www.example.com"
    host_header = "example.com"
  }
}