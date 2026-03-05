resource "sakura_simple_notification_destination" "foobar" {
    name = "foobar"
    description = "description"
    tags = ["foo","bar"]
    type = "email"
    value ="foobar@example.com"
}