data "sakura_workflows_subscription" "foobar" {}

resource "sakura_workflows" "foobar" {
  subscription_id = data.sakura_workflows_subscription.foobar.id
  name            = "sample-workflow"
  description     = "description"
  publish         = false
  logging         = false
  tags            = ["tag1", "tag2"]

  latest_revision = {
    runbook = yamlencode({
      meta = {
        description = "サンプルワークフロー"
      }
      args = {
        sample = {
          type        = "number"
          description = "サンプル引数"
        }
      }
      steps = {
        result = {
          return = "$${args.sample * 2}"
        }
      }
    })
  }
}

