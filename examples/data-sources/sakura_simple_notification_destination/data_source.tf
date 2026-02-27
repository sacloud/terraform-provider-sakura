data "sakura_simple_notification_destination" "foobar" {
  name = "foobar"
  description = "foobar"
  tags = ["foo", "bar"]
  icon_id = ""
  type = "email"
  value = "foobar@foobar.com"
}