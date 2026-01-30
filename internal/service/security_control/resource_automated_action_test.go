// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraSecurityControlAutomatedAction_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_SERVICE_PRINCIPAL_ID", "SAKURA_SIMPLE_NOTIFICATION_GROUP_ID")

	resourceName := "sakura_security_control_automated_action.foobar"
	rand := test.RandomName()
	id := os.Getenv("SAKURA_SERVICE_PRINCIPAL_ID")
	sngId := os.Getenv("SAKURA_SIMPLE_NOTIFICATION_GROUP_ID")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecurityControlAutomatedAction_basic, rand, id, sngId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "foobar-action"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "action.type", "simple_notification"),
					resource.TestCheckResourceAttr(resourceName, "action.parameters.service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName, "action.parameters.target_id", sngId),
					resource.TestCheckResourceAttr(resourceName, "execution_condition", "event.evaluationResult.status == \"REJECTED\""),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecurityControlAutomatedAction_update, rand, id, sngId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "foobar-action-updated"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "action.type", "simple_notification"),
					resource.TestCheckResourceAttr(resourceName, "action.parameters.service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName, "action.parameters.target_id", sngId),
					resource.TestCheckResourceAttr(resourceName, "execution_condition", "event.evaluationResult.evaluation_rule_id == \"disk-encryption-enabled\""),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

const testAccSakuraSecurityControlAutomatedAction_basic = `
resource "sakura_security_control_automated_action" "foobar" {
  name        = "{{ .arg0 }}"
  description = "foobar-action"
  enabled     = true
  action = {
    type = "simple_notification"
    parameters = {
      service_principal_id = "{{ .arg1 }}",
      target_id = "{{ .arg2 }}"
    }
  }
  execution_condition = "event.evaluationResult.status == \"REJECTED\""
}`

const testAccSakuraSecurityControlAutomatedAction_update = `
resource "sakura_security_control_automated_action" "foobar" {
  name        = "{{ .arg0 }}-updated"
  description = "foobar-action-updated"
  enabled     = true
  action = {
    type = "simple_notification"
    parameters = {
      service_principal_id = "{{ .arg1 }}",
      target_id = "{{ .arg2 }}"
    }
  }
  execution_condition = "event.evaluationResult.evaluation_rule_id == \"disk-encryption-enabled\""
}`

// TODO: Test with workflows after workflows resources are supported
