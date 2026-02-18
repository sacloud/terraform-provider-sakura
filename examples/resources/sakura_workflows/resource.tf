resource "sakura_workflows" "foobar" {
  name        = "sample-workflow"
  description = "description"
  publish     = true
  logging     = true
  tags        = ["tag1", "tag2"]

  revisions = [
    {
      runbook = <<-EOF
meta:
  description: サンプルワークフローv1
args:
  sample:
    type: number
    description: サンプル引数
steps:
  result:
    return: $${args.sample}
EOF
    },
    {
      alias = "v2"
      runbook = yamlencode({
        meta = {
          description = "サンプルワークフローv2"
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
    },
  ]
}
