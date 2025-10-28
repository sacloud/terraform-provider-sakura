resource "sakura_packet_filter" "foobar" {
  name        = "foobar"
  description = "description"

  expression = [{
    protocol         = "tcp"
    destination_port = "22"
  },
  {
    protocol         = "tcp"
    destination_port = "80"
  },
  {
    protocol         = "tcp"
    destination_port = "443"
  },
  {
    protocol = "icmp"
  },
  {
    protocol = "fragment"
  },
  {
    protocol    = "udp"
    source_port = "123"
  },
  {
    protocol         = "tcp"
    destination_port = "32768-61000"
  },
  {
    protocol         = "udp"
    destination_port = "32768-61000"
  },
  {
    protocol    = "ip"
    allow       = false
    description = "Deny ALL"
  }]
}