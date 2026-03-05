data "sakura_workflows_plan" "foobar" {
  name = "200Kプラン"
}

resource "sakura_workflows_subscription" "foobar" {
  plan_id = data.sakura_workflows_plan.foobar.id
}
