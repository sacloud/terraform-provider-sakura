data "sakura_archive" "ubuntu" {
  os_type = "ubuntu"
}

resource "sakura_disk" "foobar" {
  name              = "foobar"
  plan              = "ssd"
  connector         = "virtio"
  size              = 20
  source_archive_id = data.sakura_archive.ubuntu.id
  #distant_from      = ["111111111111"]

  # For encryption
  #encryption_algorithm = "aes256_xts"
  #kms_key_id           = data.sakura_kms.foobar.id

  description = "description"
  tags        = ["tag1", "tag2"]
}
