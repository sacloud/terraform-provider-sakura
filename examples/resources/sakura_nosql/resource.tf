data "sakura_switch" "foobar" {
  name = "foobar"
}

resource "sakura_nosql" "foobar" {
  name        = "foobar"
  tags        = ["nosql"]
  description = "KVS database"
  password    = "password-123456789"
  switch_id   = data.sakura_switch.foobar.id
  settings = {
    reserve_ip_address = "192.168.0.6"
    backup = {
      connect = "nfs://192.168.0.30/export"
      days_of_week = ["sun"]
      time = "00:00"
      rotate = 5
    }
    repair = {
      full = {
        interval = 14
        day_of_week = "wed"
        time = "00:30"
      }
      incremental = {
        days_of_week = ["mon"]
        time = "00:15"
      }
    }
  }
  remark = {
    nosql = {
      zone = "tk1b"
      default_user = "testuser"
    }
    servers = [
      "192.168.0.3",
      "192.168.0.4",
      "192.168.0.5",
    ]
    network = {
      gateway = "192.168.0.1"
      netmask = 24
    }
  }
  parameters = {
    concurrent_writes = "16"
    cas_contention_timeout = "2000ms"
  }
}
