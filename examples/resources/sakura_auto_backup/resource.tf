resource "sakura_disk" "foobar" {
  name = "foobar"
}

resource "sakura_auto_backup" "foobar" {
  name           = "foobar"
  disk_id        = sakura_disk.foobar.id
  days_of_week   = ["mon", "tue", "wed", "thu", "fri", "sat", "sun"]
  max_backup_num = 5
  description    = "description"
  tags           = ["tag1", "tag2"]
}
