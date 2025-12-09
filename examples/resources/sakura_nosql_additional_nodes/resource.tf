data "sakura_vswitch" "foobar" {
  name = "foobar"
}

data "sakura_nosql" "primary" {
  name = "foobar"
}

resource "sakura_nosql_additional_nodes" "foobar" {
  name        = "foobar-additional"
  tags        = ["nosql"]
  description = "KVS database additional nodes"
  vswitch_id  = data.sakura_vswitch.foobar.id
  primary_node_id = data.sakura_nosql.primary.id
  zone = data.sakura_nosql.primary.remark.nosql.zone
  settings = {
    reserve_ip_address = "192.168.0.9"
  }
  remark = {
    servers = [
      "192.168.0.7",
      "192.168.0.8",
    ]
    network = {
      gateway = "192.168.0.1"
      netmask = 24
    }
  }
}
