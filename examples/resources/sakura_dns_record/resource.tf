resource "sakura_dns" "foobar" {
  zone = "example.com"
}

resource "sakura_dns_record" "record1" {
  dns_id = sakura_dns.foobar.id
  name   = "www"
  type   = "A"
  value  = "192.168.0.1"
}

resource "sakura_dns_record" "record2" {
  dns_id = sakura_dns.foobar.id
  name   = "www"
  type   = "A"
  value  = "192.168.0.2"
}
