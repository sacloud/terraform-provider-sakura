// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraSecurityControlEvaluationRule_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_SERVICE_PRINCIPAL_ID")

	resourceName1 := "sakura_security_control_evaluation_rule.foobar1"
	resourceName2 := "sakura_security_control_evaluation_rule.foobar2"
	resourceName3 := "sakura_security_control_evaluation_rule.foobar3"
	id := os.Getenv("SAKURA_SERVICE_PRINCIPAL_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecurityControlEvaluationRule_basic, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "id", "server-no-public-ip"),
					resource.TestCheckResourceAttr(resourceName1, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName1, "parameters.service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName1, "parameters.targets.#", "2"),
					resource.TestCheckResourceAttr(resourceName1, "parameters.targets.0", "is1a"),
					resource.TestCheckResourceAttr(resourceName1, "parameters.targets.1", "tk1b"),
					resource.TestCheckResourceAttr(resourceName1, "iam_roles_required.#", "1"),
					resource.TestCheckResourceAttr(resourceName1, "no_action_on_delete", "false"),
					resource.TestCheckResourceAttr(resourceName2, "id", "elb-logging-enabled"),
					resource.TestCheckResourceAttr(resourceName2, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName2, "parameters.service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName2, "iam_roles_required.#", "1"),
					resource.TestCheckResourceAttr(resourceName2, "no_action_on_delete", "false"),
					resource.TestCheckResourceAttr(resourceName3, "id", "addon-threat-detections"),
					resource.TestCheckResourceAttr(resourceName3, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName3, "iam_roles_required.#", "1"),
					resource.TestCheckResourceAttr(resourceName3, "no_action_on_delete", "false"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecurityControlEvaluationRule_update, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "id", "server-no-public-ip"),
					resource.TestCheckResourceAttr(resourceName1, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName1, "parameters.service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName1, "parameters.targets.#", "0"),
					resource.TestCheckResourceAttr(resourceName1, "iam_roles_required.#", "1"),
					resource.TestCheckResourceAttr(resourceName1, "no_action_on_delete", "true"),
					resource.TestCheckResourceAttr(resourceName2, "id", "elb-logging-enabled"),
					resource.TestCheckResourceAttr(resourceName2, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName2, "iam_roles_required.#", "1"),
					resource.TestCheckResourceAttr(resourceName2, "no_action_on_delete", "true"),
					resource.TestCheckResourceAttr(resourceName3, "id", "addon-threat-detections"),
					resource.TestCheckResourceAttr(resourceName3, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName3, "iam_roles_required.#", "1"),
					resource.TestCheckResourceAttr(resourceName3, "no_action_on_delete", "true"),
				),
			},
		},
	})
}

const testAccSakuraSecurityControlEvaluationRule_basic = `
resource "sakura_security_control_evaluation_rule" "foobar1" {
  id         = "server-no-public-ip"
  enabled    = true
  parameters = {
    service_principal_id = "{{ .arg0 }}"
    targets = ["is1a", "tk1b"]
  }
  no_action_on_delete = false
}

resource "sakura_security_control_evaluation_rule" "foobar2" {
  id         = "elb-logging-enabled"
  enabled    = true
  parameters = {
    service_principal_id = "{{ .arg0 }}"
  }
  no_action_on_delete = false
}

resource "sakura_security_control_evaluation_rule" "foobar3" {
  id      = "addon-threat-detections"
  enabled = true
  no_action_on_delete = false
}
`

const testAccSakuraSecurityControlEvaluationRule_update = `
resource "sakura_security_control_evaluation_rule" "foobar1" {
  id         = "server-no-public-ip"
  enabled    = true
  parameters = {
    service_principal_id = "{{ .arg0 }}"
    targets = []
  }
  no_action_on_delete = true
}

resource "sakura_security_control_evaluation_rule" "foobar2" {
  id         = "elb-logging-enabled"
  enabled    = false
  parameters = {
    service_principal_id = "{{ .arg0 }}"
  }
  no_action_on_delete = true
}

resource "sakura_security_control_evaluation_rule" "foobar3" {
  id      = "addon-threat-detections"
  enabled = false
  no_action_on_delete = true
}`
