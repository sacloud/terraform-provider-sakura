data "sakura_apigw_plan" "foobar" { 
  name = "エンタープライズ"
}

resource "sakura_apigw_subscription" "foobar" {
  name = "foobar"
  plan_id = data.sakura_apigw_plan.foobar.id
}
