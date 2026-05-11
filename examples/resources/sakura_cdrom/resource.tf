resource "sakura_cdrom" "foobar" {
  name        = "foobar"
  description = "description"
  tags        = ["tag1", "tag2"]

  size           = 5
  iso_image_file = "test/dummy.iso"
}
