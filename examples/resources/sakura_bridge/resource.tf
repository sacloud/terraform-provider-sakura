resource "sakura_switch" "is1a" {
  name        = "is1a"
  description = "description"
  bridge_id   = sakura_bridge.foobar.id
  zone        = "is1a"
}

resource "sakura_switch" "is1b" {
  name        = "is1b"
  description = "description"
  bridge_id   = sakura_bridge.foobar.id
  zone        = "is1b"
}

resource "sakura_bridge" "foobar" {
  name        = "foobar"
  description = "description"
}
