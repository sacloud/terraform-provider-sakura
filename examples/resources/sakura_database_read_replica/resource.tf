data "sakura_database" "master" {
  name = "master-database-name"
}

resource "sakura_kms" "foobar" {
  name = "foobar"
}

resource "sakura_database_read_replica" "foobar" {
  master_id   = data.sakura_database.master.id
  network_interface = {
    ip_address = "192.168.11.111"
  }

  disk = {
    encryption_algorithm = "aes256_xts"
    kms_key_id           = sakura_kms.foobar.id
  }

  name        = "foobar"
  description = "description"
  tags        = ["tag1", "tag2"]
}