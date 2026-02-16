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
				Config: test.BuildConfigWithArgs(testAccSakuraWorkflows_basic, rand, sampleRunbookV1Escaped, sampleRunbookV2Terraform),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "revisions.#", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "revisions.0.alias"),
					resource.TestCheckResourceAttr(resourceName, "revisions.0.runbook", sampleRunbookV1),
					resource.TestCheckResourceAttr(resourceName, "revisions.1.alias", "v2"),
					resource.TestCheckResourceAttr(resourceName, "revisions.1.runbook", sampleRunbookV2),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraWorkflows_update, rand, sampleRunbookV1Escaped, sampleRunbookV2Escaped, sampleRunbookV3Escaped),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "publish", "true"),
					resource.TestCheckResourceAttr(resourceName, "logging", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "revisions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "revisions.0.alias", "v1"),
					resource.TestCheckResourceAttr(resourceName, "revisions.0.runbook", sampleRunbookV1),
					resource.TestCheckNoResourceAttr(resourceName, "revisions.1.alias"),
					resource.TestCheckResourceAttr(resourceName, "revisions.1.runbook", sampleRunbookV2),
					resource.TestCheckResourceAttr(resourceName, "revisions.2.alias", "v3"),
					resource.TestCheckResourceAttr(resourceName, "revisions.2.runbook", sampleRunbookV3),
					resource.TestCheckResourceAttrSet(resourceName, "revisions.2.id"),
				),
			},
		},
	})
}

func TestAccSakuraResourceWorkflows_emptyRevisions(t *testing.T) {
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraWorkflows_emptyRevisions, rand),
				ExpectError: regexp.MustCompile(`at least one revision is\s+required`),
			},
		},
	})
}

func TestAccSakuraResourceWorkflows_deleteRevisionOnUpdate(t *testing.T) {
	resourceName := "sakura_workflows.foobar"
	rand := test.RandomName()

	var workflow v1.GetWorkflowOKWorkflow
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWorkflowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraWorkflows_deleteRevisionCreate, rand, sampleRunbookV1Escaped, sampleRunbookV2Escaped),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "revisions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "revisions.0.alias", "v1"),
					resource.TestCheckResourceAttr(resourceName, "revisions.0.runbook", sampleRunbookV1),
					resource.TestCheckResourceAttr(resourceName, "revisions.1.alias", "v2"),
					resource.TestCheckResourceAttr(resourceName, "revisions.1.runbook", sampleRunbookV2),
				),
			},
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraWorkflows_deleteRevisionUpdate, rand, sampleRunbookV2Escaped),
				ExpectError: regexp.MustCompile(`Deletion of existing revision is not supported, but revisions\[0\] is deleted.`),
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
				Config:      test.BuildConfigWithArgs(testAccSakuraWorkflows_invalidName, randInvalid, sampleRunbookV1Escaped),
				ExpectError: regexp.MustCompile("Invalid workflow configuration"),
			},
		},
	})
}

func TestAccSakuraResourceWorkflows_reorderRevisions(t *testing.T) {
	resourceName := "sakura_workflows.foobar"
	rand := test.RandomName()

	var workflow v1.GetWorkflowOKWorkflow
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWorkflowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraWorkflows_reorderRevisions, rand, sampleRunbookV1Escaped, sampleRunbookV2Escaped),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "revisions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "revisions.0.runbook", sampleRunbookV1),
					resource.TestCheckResourceAttr(resourceName, "revisions.1.runbook", sampleRunbookV2),
				),
			},
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraWorkflows_reorderRevisions, rand, sampleRunbookV2Escaped, sampleRunbookV1Escaped),
				ExpectError: regexp.MustCompile("Reordering revisions is not supported"),
			},
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraWorkflows_insertRevisions, rand, sampleRunbookV1Escaped, sampleRunbookV3Escaped, sampleRunbookV2Escaped),
				ExpectError: regexp.MustCompile("Reordering revisions is not supported"),
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

		client := test.AccClientGetter()
		workflowOp := workflows.NewWorkflowOp(client.WorkflowsClient)

		foundWorkflow, err := workflowOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		foundID := foundWorkflow.ID
		if foundID != rs.Primary.ID {
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

const sampleRunbookV3 = `meta:
  description: サンプルワークフローv3
args:
  sample:
    type: number
    description: サンプル引数
steps:
  result:
    return: ${args.sample * 3}
`

var (
	sampleRunbookV1Escaped = strings.ReplaceAll(sampleRunbookV1, "$", "$$")
	sampleRunbookV2Escaped = strings.ReplaceAll(sampleRunbookV2, "$", "$$")
	sampleRunbookV3Escaped = strings.ReplaceAll(sampleRunbookV3, "$", "$$")
)

var testAccSakuraWorkflows_basic = `
resource "sakura_workflows" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  publish = false
  logging = false
  tags    = ["tag1", "tag2"]

  revisions = [
    {
      runbook = <<-EOF
{{ .arg1 }}EOF
    },
    {
      alias   = "v2"
      runbook = yamlencode({{ .arg2 }})
    },
  ]
}`

var testAccSakuraWorkflows_update = `
resource "sakura_workflows" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  publish = true
  logging = true
  tags    = ["tag2"]

  revisions = [
    {
      alias   = "v1"
      runbook = <<-EOF
{{ .arg1 }}EOF
    },
    {
      runbook = <<-EOF
{{ .arg2 }}EOF
    },
    {
      alias   = "v3"
      runbook = <<-EOF
{{ .arg3 }}EOF
    },
  ]
}`

var testAccSakuraWorkflows_emptyRevisions = `
resource "sakura_workflows" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  publish = false
  logging = false

  revisions = []
}`

var testAccSakuraWorkflows_deleteRevisionCreate = `
resource "sakura_workflows" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  publish = false
  logging = false

  revisions = [
    {
      alias   = "v1"
      runbook = <<-EOF
{{ .arg1 }}EOF
    },
    {
      alias   = "v2"
      runbook = <<-EOF
{{ .arg2 }}EOF
    },
  ]
}`

var testAccSakuraWorkflows_deleteRevisionUpdate = `
resource "sakura_workflows" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  publish = false
  logging = false

  revisions = [
    # deleted v1
    {
      alias   = "v2"
      runbook = <<-EOF
{{ .arg1 }}EOF
    },
  ]
}`

var testAccSakuraWorkflows_invalidName = `
resource "sakura_workflows" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  publish = false
  logging = false

  revisions = [
    {
      runbook = <<-EOF
{{ .arg1 }}EOF
    },
  ]
}`

var testAccSakuraWorkflows_reorderRevisions = `
resource "sakura_workflows" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  publish = false
  logging = false

  revisions = [
    {
      runbook = <<-EOF
{{ .arg1 }}EOF
    },
    {
      runbook = <<-EOF
{{ .arg2 }}EOF
    },
  ]
}`

var testAccSakuraWorkflows_insertRevisions = `
resource "sakura_workflows" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  publish = false
  logging = false

  # insert new revision in the middle
  revisions = [
    {
      runbook = <<-EOF
{{ .arg1 }}EOF
    },
    {
      runbook = <<-EOF
{{ .arg2 }}EOF
    },
    {
      runbook = <<-EOF
{{ .arg3 }}EOF
    },
  ]
}`
