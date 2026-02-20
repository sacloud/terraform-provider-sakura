// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
	"github.com/sacloud/workflows-api-go"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

func TestAccSakuraResourceWorkflowsRevisionAlias_basic(t *testing.T) {
	resourceName := "sakura_workflows_revision_alias.foobar"
	rand := test.RandomName()

	var workflow v1.GetWorkflowOKWorkflow
	var revisionAlias v1.GetWorkflowRevisionsOKRevision

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWorkflowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraWorkflowsRevisionAlias_basic, rand, sampleRunbookV2Terraform),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists("sakura_workflows.foobar", &workflow),
					testCheckSakuraWorkflowsRevisionAliasExists(resourceName, &revisionAlias),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_id", "sakura_workflows.foobar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "revision_id", "sakura_workflows.foobar", "latest_revision.id"),
					resource.TestCheckResourceAttr(resourceName, "alias", "stable"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraWorkflowsRevisionAlias_update, rand, sampleRunbookV2Terraform),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists("sakura_workflows.foobar", &workflow),
					testCheckSakuraWorkflowsRevisionAliasExists(resourceName, &revisionAlias),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_id", "sakura_workflows.foobar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "revision_id", "sakura_workflows.foobar", "latest_revision.id"),
					resource.TestCheckResourceAttr(resourceName, "alias", "stable-updated"),
				),
			},
		},
	})
}

func TestAccSakuraResourceWorkflowsRevisionAlias_validation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWorkflowsDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccSakuraWorkflowsRevisionAlias_emptyWorkflowID,
				ExpectError: regexp.MustCompile(`Attribute workflow_id string length must be at least 1, got: 0`),
			},
			{
				Config:      testAccSakuraWorkflowsRevisionAlias_emptyRevisionID,
				ExpectError: regexp.MustCompile(`The argument "revision_id" is required, but no definition was found.`),
			},
			{
				Config:      testAccSakuraWorkflowsRevisionAlias_invalidRevisionID,
				ExpectError: regexp.MustCompile(`Attribute revision_id needs to be a string representation of an integer.`),
			},
			{
				Config:      testAccSakuraWorkflowsRevisionAlias_invalidAlias,
				ExpectError: regexp.MustCompile(`Revision alias validation failed:`),
			},
		},
	})
}

func TestAccSakuraResourceWorkflowsRevisionAlias_workflowUpdate(t *testing.T) {
	resourceName := "sakura_workflows_revision_alias.foobar"
	workflowResourceName := "sakura_workflows.foobar"
	rand := test.RandomName()

	var workflow v1.GetWorkflowOKWorkflow
	var revisionAlias v1.GetWorkflowRevisionsOKRevision

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWorkflowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraWorkflowsRevisionAlias_workflowUpdate_step1, rand, sampleRunbookV1Escaped),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists(workflowResourceName, &workflow),
					testCheckSakuraWorkflowsRevisionAliasExists(resourceName, &revisionAlias),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_id", workflowResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "revision_id", workflowResourceName, "latest_revision.id"),
					resource.TestCheckResourceAttr(resourceName, "alias", "stable"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraWorkflowsRevisionAlias_workflowUpdate_step2, rand, sampleRunbookV2Terraform),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists(workflowResourceName, &workflow),
					testCheckSakuraWorkflowsRevisionAliasExists(resourceName, &revisionAlias),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_id", workflowResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "revision_id", workflowResourceName, "latest_revision.id"),
					resource.TestCheckResourceAttr(resourceName, "alias", "stable"),
				),
			},
		},
	})
}

func testCheckSakuraWorkflowsRevisionAliasExists(resourceName string, revisionAlias *v1.GetWorkflowRevisionsOKRevision) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		workflowID := rs.Primary.Attributes["workflow_id"]
		revisionIDstr := rs.Primary.Attributes["revision_id"]
		alias := rs.Primary.Attributes["alias"]

		if workflowID == "" {
			return fmt.Errorf("workflow_id is not set")
		}
		if revisionIDstr == "" {
			return fmt.Errorf("revision_id is not set")
		}
		if alias == "" {
			return fmt.Errorf("alias is not set")
		}

		revisionID, err := strconv.Atoi(revisionIDstr)
		if err != nil {
			return fmt.Errorf("failed to parse revisionID: %s", err)
		}

		client := test.AccClientGetter()
		revisionOp := workflows.NewRevisionOp(client.WorkflowsClient)

		rev, err := revisionOp.Read(context.Background(), workflowID, revisionID)
		if err != nil {
			return fmt.Errorf("failed to read workflow revision: %s", err)
		}

		revAlias, ok := rev.RevisionAlias.Get()
		if !ok {
			return errors.New("revision alias is not set")
		}

		if revAlias != alias {
			return fmt.Errorf("expected alias %s, got %s", alias, revAlias)
		}

		if revisionAlias != nil {
			*revisionAlias = *rev
		}

		return nil
	}
}

const testAccSakuraWorkflowsRevisionAlias_basic = `
data "sakura_workflows_plan" "foobar" {
  name = "200K"
}

resource "sakura_workflows_subscription" "foobar" {
  plan_id = data.sakura_workflows_plan.foobar.id
}

resource "sakura_workflows" "foobar" {
  subscription_id = sakura_workflows_subscription.foobar.id
  name            = "{{ .arg0 }}"
  publish         = false
  logging         = false

  latest_revision = {
    runbook = yamlencode({{ .arg1 }})
  }
}

resource "sakura_workflows_revision_alias" "foobar" {
  workflow_id = sakura_workflows.foobar.id
  revision_id = sakura_workflows.foobar.latest_revision.id
  alias       = "stable"
}
`

const testAccSakuraWorkflowsRevisionAlias_update = `
data "sakura_workflows_plan" "foobar" {
  name = "200K"
}

resource "sakura_workflows_subscription" "foobar" {
  plan_id = data.sakura_workflows_plan.foobar.id
}

resource "sakura_workflows" "foobar" {
  subscription_id = sakura_workflows_subscription.foobar.id
  name            = "{{ .arg0 }}"
  publish         = false
  logging         = false

  latest_revision = {
    runbook = yamlencode({{ .arg1 }})
  }
}

resource "sakura_workflows_revision_alias" "foobar" {
  workflow_id = sakura_workflows.foobar.id
  revision_id = sakura_workflows.foobar.latest_revision.id
  alias       = "stable-updated"
}
`

const testAccSakuraWorkflowsRevisionAlias_emptyWorkflowID = `
resource "sakura_workflows_revision_alias" "foobar" {
  workflow_id = ""
  revision_id = "1"
  alias       = "foobar"
}
`

const testAccSakuraWorkflowsRevisionAlias_emptyRevisionID = `
resource "sakura_workflows_revision_alias" "foobar" {
  workflow_id   = "foobar"
	# revision_id =
  alias         = "foobar"
}
`

const testAccSakuraWorkflowsRevisionAlias_invalidRevisionID = `
resource "sakura_workflows_revision_alias" "foobar" {
  workflow_id = "foobar"
	revision_id = "not-a-number"
  alias       = "foobar"
}
`

const testAccSakuraWorkflowsRevisionAlias_invalidAlias = `
resource "sakura_workflows_revision_alias" "foobar" {
  workflow_id = "foobar"
  revision_id = "1"
  alias       = "./+_-"
}
`

const testAccSakuraWorkflowsRevisionAlias_workflowUpdate_step1 = `
data "sakura_workflows_plan" "foobar" {
  name = "200K"
}

resource "sakura_workflows_subscription" "foobar" {
  plan_id = data.sakura_workflows_plan.foobar.id
}

resource "sakura_workflows" "foobar" {
  subscription_id = sakura_workflows_subscription.foobar.id
  name            = "{{ .arg0 }}"
  publish         = false
  logging         = false

  latest_revision = {
    runbook = <<-EOF
{{ .arg1 }}EOF
  }
}

resource "sakura_workflows_revision_alias" "foobar" {
  workflow_id = sakura_workflows.foobar.id
  revision_id = sakura_workflows.foobar.latest_revision.id
  alias       = "stable"
}
`

const testAccSakuraWorkflowsRevisionAlias_workflowUpdate_step2 = `
data "sakura_workflows_plan" "foobar" {
  name = "200K"
}

resource "sakura_workflows_subscription" "foobar" {
  plan_id = data.sakura_workflows_plan.foobar.id
}

resource "sakura_workflows" "foobar" {
  subscription_id = sakura_workflows_subscription.foobar.id
  name            = "{{ .arg0 }}"
  publish         = false
  logging         = false

  latest_revision = {
    runbook = yamlencode({{ .arg1 }})
  }
}

resource "sakura_workflows_revision_alias" "foobar" {
  workflow_id = sakura_workflows.foobar.id
  revision_id = sakura_workflows.foobar.latest_revision.id
  alias       = "stable"
}
`
