data "sakura_vswitch" "foobar" {
  name = "foobar"
}

resource "sakura_nosql" "foobar" {
  name        = "foobar"
  tags        = ["nosql"]
  zone        = "tk1b"
  plan        = "100GB" // or "250GB"
  description = "NoSQL database"
  password    = "password-123456789"
  vswitch_id  = data.sakura_vswitch.foobar.id
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
      default_user = "testuser"
      port = 9042
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
  /* sakura_kms based disk encryption
  disk = {
    encryption_algorithm = "aes256_xts"
    kms_key_id           = data.sakura_kms.foobar.id
  }
   */
  parameters = {
    concurrent_writes = "16"
    cas_contention_timeout = "2000ms"
  }
}

resource "sakura_nosql" "foobar40GB" {
  name = "foobar-40GB"
  tags = ["nosql"]
  plan = "40GB"
  description = "Test database"
  password    = "sdktest-12345678"
  vswitch_id  = data.sakura_vswitch.foobar.id
  settings = {
    backup = {
      connect = "nfs://192.168.0.31/export"
      days_of_week = ["sun"]
      time = "00:30"
      rotate = 2
    }
  }
  remark = {
    nosql = {
      default_user = "testuser2"
      port = 9042
    }
    servers = [
      "192.168.0.7",
    ]
    network = {
      gateway = "192.168.0.1"
      netmask = 24
    }
  }
}