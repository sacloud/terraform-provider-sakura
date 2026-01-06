variable username {}
variable password {}
variable replica_password {}

resource "sakura_database" "foobar" {
  database_type    = "mariadb"
  database_version = "10.11" // optional
  plan             = "30g"

  username = var.username
  password_wo = var.password
  replica_password_wo = var.replica_password
  password_wo_version = 1
  // for backward compatibility
  //password = var.password
  //replica_password = var.replica_password

  network_interface = {
    vswitch_id    = sakura_vswitch.foobar.id
    ip_address    = "192.168.11.11"
    netmask       = 24
    gateway       = "192.168.11.1"
    port          = 3306
    source_ranges = ["192.168.11.0/24", "192.168.12.0/24"]
  }

  backup = {
    days_of_week = ["mon", "tue"]
    time         = "00:00"
  }

  # continuous_backupを指定するときはdatabase_versionが必須
  # continuous_backup = {
  #   days_of_week = ["mon", "tue"]
  #   time         = "01:30"
  #   connect      = "nfs://${sakura_nfs.foobar.network_interface.ip_address}/export"
  # }

  parameters = {
    max_connections = 100
  }

  monitoring_suite = {
    enabled = true
  }

  disk = {
    encryption_algorithm = "aes256_xts"
    kms_key_id           = sakura_kms.foobar.id
  }

  name        = "foobar"
  description = "description"
  tags        = ["tag1", "tag2"]
}

resource "sakura_vswitch" "foobar" {
  name = "foobar"
}

resource "sakura_nfs" "foobar" {
  name = "foobar"
  plan = "ssd"
  size = "100"

  network_interface = {
    vswitch_id = sakura_vswitch.foobar.id
    ip_address = "192.168.11.111"
    netmask    = 24
    gateway    = "192.168.11.1"
  }
}

resource "sakura_kms" "foobar" {
  name = "foobar"
}
