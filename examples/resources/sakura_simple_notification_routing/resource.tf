resource "sakura_simple_notification_routing" "foobar" {
  name            = "foobar" 
  description     = "description"
  tags            = ["foo","bar"]
  icon_id         = sakura_icon.foobar.id
  match_labels    = []
  source_id       = "<source-id>" 
  target_group_id = sakura_simple_notification_group.foobar.id
}