// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

func TestAccSakuraDataSourceWorkflows_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t)
	resourceName := "data.sakura_workflows.foobar"
	rand := test.RandomName()

	var workflow v1.GetWorkflowOKWorkflow
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWorkflowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceWorkflows_basic, rand, sampleRunbookV2Terraform),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraWorkflowsExists("sakura_workflows.foobar", &workflow),
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "publish", "false"),
					resource.TestCheckResourceAttr(resourceName, "logging", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "latest_revision.runbook", sampleRunbookV2),
					resource.TestCheckResourceAttrPair(resourceName, "latest_revision.id", "sakura_workflows.foobar", "latest_revision.id"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_revision.created_at", "sakura_workflows.foobar", "latest_revision.created_at"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_revision.updated_at", "sakura_workflows.foobar", "latest_revision.updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "subscription_id", "sakura_workflows_subscription.foobar", "id"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceWorkflows_basic = `
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
    runbook = yamlencode({{ .arg1 }})
  }
}

data "sakura_workflows" "foobar" {
  subscription_id = sakura_workflows_subscription.foobar.id
  id              = sakura_workflows.foobar.id
}`
