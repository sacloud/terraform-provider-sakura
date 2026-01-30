// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceActivation_Basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_SERVICE_PRINCIPAL_ID")

	resourceName := "data.sakura_security_control_activation.foobar"
	id := os.Getenv("SAKURA_SERVICE_PRINCIPAL_ID")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceActivationConfig, id),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_principal_id", id),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "automated_action_limit"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceActivationConfig = `
resource "sakura_security_control_activation" "foobar" {
  service_principal_id = "{{ .arg0 }}"
  enabled = true
  no_action_on_delete = true
}

data "sakura_security_control_activation" "foobar" {
	depends_on = [sakura_security_control_activation.foobar]
}`
