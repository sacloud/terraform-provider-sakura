resource sakura_internet "foobar" {
  name = "foobar"
}

resource "sakura_subnet" "foobar" {
  internet_id = sakura_internet.foobar.id
  next_hop    = sakura_internet.foobar.min_ip_address
}