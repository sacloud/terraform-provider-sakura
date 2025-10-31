resource "sakura_eventbus_trigger" "foobar" {
  name        = "foobar"
  description = "description"
  tags        = ["tag1", "tag2"]

  process_configuration_id = sakura_eventbus_process_configuration.foobar.id
  source                   = "test-source"
  types                    = ["type1", "type2"]
  conditions               = [
    {
      key    = "key1"
      op     = "in"
      values = ["foo", "bar"]
    },
    {
      key    = "key2"
      op     = "eq"
      values = ["buz"]
    },
  ]
}
