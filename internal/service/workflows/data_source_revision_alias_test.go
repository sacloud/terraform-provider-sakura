// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

func TestAccSakuraDataSourceWorkflowsRevisionAlias_basic(t *testing.T) {
	resourceName := "data.sakura_workflows_revision_alias.foobar"
	rand := test.RandomName()

	var workflow v1.GetWorkflowOKWorkflow
	var revisionAlias v1.GetWorkflowRevisionsOKRevision

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWorkflowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceWorkflowsRevisionAlias_basic, rand, sampleRunbookV2Terraform),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists("sakura_workflows.foobar", &workflow),
					testCheckSakuraWorkflowsRevisionAliasExists(resourceName, &revisionAlias),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_id", "sakura_workflows.foobar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "revision_id", "sakura_workflows.foobar", "latest_revision.id"),
					resource.TestCheckResourceAttrPair(resourceName, "alias", "sakura_workflows_revision_alias.foobar", "alias"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceWorkflowsRevisionAlias_update, rand, sampleRunbookV2Terraform),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists("sakura_workflows.foobar", &workflow),
					testCheckSakuraWorkflowsRevisionAliasExists(resourceName, &revisionAlias),
					resource.TestCheckResourceAttrPair(resourceName, "workflow_id", "sakura_workflows.foobar", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "revision_id", "sakura_workflows.foobar", "latest_revision.id"),
					resource.TestCheckResourceAttrPair(resourceName, "alias", "sakura_workflows_revision_alias.foobar", "alias"),
				),
			},
		},
	})
}

const testAccSakuraDataSourceWorkflowsRevisionAlias_basic = `
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

data "sakura_workflows_revision_alias" "foobar" {
  workflow_id = sakura_workflows_revision_alias.foobar.workflow_id
  revision_id = sakura_workflows_revision_alias.foobar.revision_id
}
`

const testAccSakuraDataSourceWorkflowsRevisionAlias_update = `
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

data "sakura_workflows_revision_alias" "foobar" {
  workflow_id = sakura_workflows_revision_alias.foobar.workflow_id
  revision_id = sakura_workflows_revision_alias.foobar.revision_id
}
`
