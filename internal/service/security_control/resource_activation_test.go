// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraSecurityControlActivation_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_SERVICE_PRINCIPAL_ID")

	resourceName := "sakura_security_control_activation.foobar"
	id := os.Getenv("SAKURA_SERVICE_PRINCIPAL_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecurityControlActivation_basic, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "automated_action_limit"),
					resource.TestCheckResourceAttr(resourceName, "no_action_on_delete", "false"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSecurityControlActivation_update, id),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "automated_action_limit"),
					resource.TestCheckResourceAttr(resourceName, "no_action_on_delete", "true"),
				),
			},
		},
	})
}

const testAccSakuraSecurityControlActivation_basic = `
resource "sakura_security_control_activation" "foobar" {
  service_principal_id = "{{ .arg0 }}"
  enabled = false
  no_action_on_delete = false
}
`

const testAccSakuraSecurityControlActivation_update = `
resource "sakura_security_control_activation" "foobar" {
  service_principal_id = "{{ .arg0 }}"
  enabled = true
  no_action_on_delete = true
}`
