// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceEvaluationRule_Basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_SERVICE_PRINCIPAL_ID")

	resourceName := "data.sakura_security_control_evaluation_rule.foobar"
	id := os.Getenv("SAKURA_SERVICE_PRINCIPAL_ID")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceEvaluationRuleConfig, id),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "id", "server-no-public-ip"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameters.service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName, "parameters.targets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.targets.0", "is1c"),
					resource.TestCheckResourceAttr(resourceName, "parameters.targets.1", "tk1a"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles_required.#", "1"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceEvaluationRuleConfig = `
resource "sakura_security_control_evaluation_rule" "foobar" {
  id         = "server-no-public-ip"
  enabled    = true
  parameters = {
    service_principal_id = "{{ .arg0 }}"
    targets = ["is1c", "tk1a"]
  }
}

data "sakura_security_control_evaluation_rule" "foobar" {
  id = sakura_security_control_evaluation_rule.foobar.id
}`
