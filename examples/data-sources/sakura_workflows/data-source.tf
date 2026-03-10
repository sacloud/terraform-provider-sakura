data "sakura_workflows_subscription" "foobar" {}

data "sakura_workflows" "foobar" {
  subscription_id = data.sakura_workflows_subscription.foobar.id
  id              = "workflow-id"
}

