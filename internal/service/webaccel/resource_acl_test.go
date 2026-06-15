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

func TestAccResourceSakuraWebAccelACL_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, envWebAccelSiteName)

	siteName := os.Getenv(envWebAccelSiteName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelACLConfig(siteName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("sakura_webaccel_acl.foobar", "acl", "deny 192.0.2.5/25\ndeny 198.51.100.0\nallow all"),
				),
			},
		},
	})
}

func TestAccImportSakuraWebAccelACL_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, envWebAccelSiteName)

	siteName := os.Getenv(envWebAccelSiteName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelACLConfig(siteName),
			},
			{
				ResourceName:      "sakura_webaccel_acl.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckSakuraWebAccelACLDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := webaccel.NewOp(client.WebaccelClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_webaccel_acl" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		res, err := op.ReadACL(context.Background(), rs.Primary.ID)
		if err == nil && res.ACL != "" {
			return fmt.Errorf("still exists WebAccel ACL: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckSakuraWebAccelACLConfig(name string) string {
	tmpl := `
data "sakura_webaccel" "site" {
  name = "%s"
}
resource "sakura_webaccel_acl" "foobar" {
  site_id = data.sakura_webaccel.site.id

  acl = join("\n", [
    "deny 192.0.2.5/25",
    "deny 198.51.100.0",
    "allow all",
  ])
}
`
	return fmt.Sprintf(tmpl, name)
}
