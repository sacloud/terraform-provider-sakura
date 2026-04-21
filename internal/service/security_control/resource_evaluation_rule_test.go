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
	test.SkipIfEnvIsNotSet(t, "SAKURA_SECURITY_CONTROL_SERVICE_PRINCIPAL_ID")

	resourceName1 := "sakura_security_control_evaluation_rule.foobar1"
	resourceName2 := "sakura_security_control_evaluation_rule.foobar2"
	resourceName4 := "sakura_security_control_evaluation_rule.foobar4"
	id := os.Getenv("SAKURA_SECURITY_CONTROL_SERVICE_PRINCIPAL_ID")

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
					resource.TestCheckResourceAttr(resourceName4, "id", "objectstorage-bucket-encryption-enabled"),
					resource.TestCheckResourceAttr(resourceName4, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName4, "parameters.service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName4, "parameters.targets.#", "2"),
					resource.TestCheckResourceAttr(resourceName4, "parameters.targets.0", "isk01"),
					resource.TestCheckResourceAttr(resourceName4, "parameters.targets.1", "tky01"),
					resource.TestCheckResourceAttr(resourceName4, "iam_roles_required.#", "1"),
					resource.TestCheckResourceAttr(resourceName4, "no_action_on_delete", "false"),
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
					resource.TestCheckResourceAttr(resourceName4, "id", "objectstorage-bucket-encryption-enabled"),
					resource.TestCheckResourceAttr(resourceName4, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName4, "parameters.service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName4, "parameters.targets.#", "0"),
					resource.TestCheckResourceAttr(resourceName4, "iam_roles_required.#", "1"),
					resource.TestCheckResourceAttr(resourceName4, "no_action_on_delete", "true"),
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

resource "sakura_security_control_evaluation_rule" "foobar4" {
  id      = "objectstorage-bucket-encryption-enabled"
  enabled = true
  no_action_on_delete = false
  parameters = {
    service_principal_id = "{{ .arg0 }}"
    targets = ["isk01", "tky01"]
  }
}`

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

resource "sakura_security_control_evaluation_rule" "foobar4" {
  id      = "objectstorage-bucket-encryption-enabled"
  enabled = false
  no_action_on_delete = true
  parameters = {
    service_principal_id = "{{ .arg0 }}"
    targets = []
  }
}`

func TestAccSakuraSecurityControlEvaluationRule_addon(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_SECURITY_CONTROL_SERVICE_PRINCIPAL_ID")
	test.SkipIfEnvIsNotSet(t, "SAKURA_ENABLE_ADDON_TEST")

	resourceName := "sakura_security_control_evaluation_rule.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSakuraSecurityControlEvaluationRule_addon_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "addon-threat-detections"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles_required.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "no_action_on_delete", "false"),
				),
			},
			{
				Config: testAccSakuraSecurityControlEvaluationRule_addon_update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "addon-threat-detections"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles_required.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "no_action_on_delete", "true"),
				),
			},
		},
	})
}

const testAccSakuraSecurityControlEvaluationRule_addon_basic = `
resource "sakura_security_control_evaluation_rule" "foobar" {
  id      = "addon-threat-detections"
  enabled = true
  no_action_on_delete = false
}`

const testAccSakuraSecurityControlEvaluationRule_addon_update = `
resource "sakura_security_control_evaluation_rule" "foobar" {
  id      = "addon-threat-detections"
  enabled = false
  no_action_on_delete = true
}`
