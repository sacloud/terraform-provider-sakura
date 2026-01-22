data "sakura_apigw_service" "foobar" {
  name = "foobar"
}

data "sakura_apigw_route" "foobar" {
  name = "foobar"
  service_id = data.sakura_apigw_service.foobar.id
}
