resource "sakura_dns" "foobar" {
  zone        = "example.com"
  description = "description"
  tags        = ["tag1", "tag2"]
  monitoring_suite = {
    enabled = true
  }
}
