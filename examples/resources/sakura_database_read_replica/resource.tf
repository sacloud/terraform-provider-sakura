resource "sakura_database_read_replica" "foobar" {
  master_id   = data.sakura_database.master.id
  network_interface = {
    ip_address  = "192.168.11.111"
  }
  name        = "foobar"
  description = "description"
  tags        = ["tag1", "tag2"]
}

data "sakura_database" "master" {
  name = "master-database-name"
}
