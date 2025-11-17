data "sakura_switch" "foobar" {
  name = "foobar"
}

data "sakura_nosql" "primary" {
  name = "foobar"
}

resource "sakura_nosql_additional_nodes" "foobar" {
  name        = "foobar-additional"
  tags        = ["nosql"]
  description = "KVS database additional nodes"
  switch_id   = data.sakura_switch.foobar.id
  settings = {
    reserve_ip_address = "192.168.0.9"
  }
  remark = {
    nosql = {
      primary_nodes = {
        id = data.sakura_nosql.primary.id
        zone = data.sakura_nosql.primary.remark.nosql.zone
      }
      version = data.sakura_nosql.primary.remark.nosql.version
      zone = data.sakura_nosql.primary.remark.nosql.zone
    }
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
