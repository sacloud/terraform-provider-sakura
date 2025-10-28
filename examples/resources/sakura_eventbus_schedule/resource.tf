resource "sakura_eventbus_schedule" "foobar" {
  name        = "foobar"
  description = "description"
  tags        = ["tag1", "tag2"]

  process_configuration_id = sakura_eventbus_process_configuration.foobar.id
  starts_at                = 1700000000000
  recurring_step           = 1
  recurring_unit           = "day"
  # or
  # crontab                  = "*/15 * * * *"
}