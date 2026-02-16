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
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceWorkflows_basic, rand, sampleRunbookV1Escaped, sampleRunbookV2Terraform),
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
					resource.TestCheckNoResourceAttr(resourceName, "revisions.0.alias"),
					resource.TestCheckResourceAttr(resourceName, "revisions.0.runbook", sampleRunbookV1),
					resource.TestCheckResourceAttr(resourceName, "revisions.1.alias", "v2"),
					resource.TestCheckResourceAttr(resourceName, "revisions.1.runbook", sampleRunbookV2),
				),
			},
		},
	})
}

var testAccSakuraDataSourceWorkflows_basic = `
resource "sakura_workflows" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  publish     = false
  logging     = false
  tags        = ["tag1", "tag2"]

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
}

data "sakura_workflows" "foobar" {
  id = sakura_workflows.foobar.id
}`
