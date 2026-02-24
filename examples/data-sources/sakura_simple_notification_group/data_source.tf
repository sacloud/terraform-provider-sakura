data "sakura_simple_notification_group" "foobar" {
  name = "foobar"
  description = "foobar"
  tags = ["foo", "bar"]
  destinations = [sakura_simple_notification_destination.id]
}