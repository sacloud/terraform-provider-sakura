// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package webaccel_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
	"github.com/sacloud/webaccel-api-go"
)

func TestAccResourceSakuraWebAccelActivation_Basic(t *testing.T) {
	envKeys := []string{
		envWebAccelSiteName,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is required. skip", k)
			return
		}
	}

	siteName := os.Getenv(envWebAccelSiteName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelActivationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelActivationConfig(siteName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel_activation.foobar", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceSakuraWebAccelActivation_Update(t *testing.T) {
	envKeys := []string{
		envWebAccelSiteName,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is required. skip", k)
			return
		}
	}

	siteName := os.Getenv(envWebAccelSiteName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelActivationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelActivationConfig(siteName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel_activation.foobar", "enabled", "true"),
				),
			},
			{
				Config: testAccCheckSakuraWebAccelActivationConfig(siteName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel_activation.foobar", "enabled", "false"),
				),
			},
			{
				Config: testAccCheckSakuraWebAccelActivationConfig(siteName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel_activation.foobar", "enabled", "true"),
				),
			},
		},
	})
}

func testCheckSakuraWebAccelActivationDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := webaccel.NewOp(client.WebaccelClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_webaccel_activation" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		res, err := op.Read(context.Background(), rs.Primary.ID)
		if err == nil && res.Status == "enabled" {
			return fmt.Errorf("still exists WebAccel Activation: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckSakuraWebAccelActivationConfig(name string, status bool) string {
	statusValue := "false"
	if status {
		statusValue = "true"
	}
	tmpl := `
data "sakura_webaccel" "site" {
  name = "%s"
}
resource "sakura_webaccel_activation" "foobar" {
  site_id = data.sakura_webaccel.site.id
  enabled = %s
}
`
	return fmt.Sprintf(tmpl, name, statusValue)
}
