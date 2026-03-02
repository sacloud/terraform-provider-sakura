// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
	"github.com/sacloud/workflows-api-go"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

func TestAccSakuraResourceWorkflows_basic(t *testing.T) {
	resourceName := "sakura_workflows.foobar"
	rand := test.RandomName()

	var workflow v1.GetWorkflowOKWorkflow
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWorkflowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraWorkflows_basic, rand, sampleRunbookV1Escaped),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttrPair(resourceName, "subscription_id", "sakura_workflows_subscription.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision.runbook", sampleRunbookV1),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision.id"),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision.updated_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraWorkflows_update, rand, sampleRunbookV2Terraform),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "logging", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag2"),
					resource.TestCheckResourceAttrPair(resourceName, "subscription_id", "sakura_workflows_subscription.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision.runbook", sampleRunbookV2),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision.id"),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision.created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "latest_revision.updated_at"),
				),
			},
		},
	})
}

func TestAccSakuraResourceWorkflows_invalidName(t *testing.T) {
	randInvalid := test.RandomName() + "./+invalid"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraWorkflows_basic, randInvalid, sampleRunbookV1Escaped),
				ExpectError: regexp.MustCompile("Invalid workflow configuration"),
			},
		},
	})
}

func testCheckSakuraWorkflowsDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	workflowOp := workflows.NewWorkflowOp(client.WorkflowsClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_workflows" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := workflowOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists Workflow: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraWorkflowsExists(n string, workflow *v1.GetWorkflowOKWorkflow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Workflow ID is set")
		}

		workflowOp := workflows.NewWorkflowOp(test.AccClientGetter().WorkflowsClient)

		foundWorkflow, err := workflowOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundWorkflow.ID != rs.Primary.ID {
			return fmt.Errorf("not found Workflow: %s", rs.Primary.ID)
		}

		*workflow = *foundWorkflow
		return nil
	}
}

const sampleRunbookV1 = `meta:
  description: サンプルワークフローv1
args:
  sample:
    type: number
    description: サンプル引数
steps:
  result:
    return: ${args.sample}
`

// NOTE: yamlencode would sort the keys in alphabetical order.
const sampleRunbookV2 = `"args":
  "sample":
    "description": "サンプル引数"
    "type": "number"
"meta":
  "description": "サンプルワークフローv2"
"steps":
  "result":
    "return": "${args.sample * 2}"
`

const sampleRunbookV2Terraform = `{
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
}`

var sampleRunbookV1Escaped = strings.ReplaceAll(sampleRunbookV1, "$", "$$")

var testAccSakuraWorkflows_basic = `
data "sakura_workflows_plan" "foobar" {
  name = "200K"
}

resource "sakura_workflows_subscription" "foobar" {
  plan_id = data.sakura_workflows_plan.foobar.id
}

resource "sakura_workflows" "foobar" {
  subscription_id = sakura_workflows_subscription.foobar.id
  name            = "{{ .arg0 }}"
  description     = "description"
  publish         = false
  logging         = false
  tags            = ["tag1", "tag2"]

  latest_revision = {
    runbook = <<-EOF
{{ .arg1 }}EOF
  }
}`

var testAccSakuraWorkflows_update = `
data "sakura_workflows_plan" "foobar" {
  name = "200K"
}

resource "sakura_workflows_subscription" "foobar" {
  plan_id = data.sakura_workflows_plan.foobar.id
}

resource "sakura_workflows" "foobar" {
  subscription_id = sakura_workflows_subscription.foobar.id
  name            = "{{ .arg0 }}"
  description     = "description-updated"
  publish         = true
  logging         = true
  tags            = ["tag2"]

  latest_revision = {
    runbook = yamlencode({{ .arg1 }})
  }
}`
