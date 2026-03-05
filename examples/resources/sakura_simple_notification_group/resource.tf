resource "sakura_simple_notification_group" "foobar" {
    name = "foobar"
    description = "description"
    tags = ["foo","bar"]
    destinations = [sakura_simple_notification_destination.foobar.id]
}