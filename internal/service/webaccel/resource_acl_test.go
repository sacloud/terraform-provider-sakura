// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package webaccel_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccResourceSakuraWebAccelACL_basic(t *testing.T) {
	envKeys := []string{
		envWebAccelSiteName,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is requilred. skip", k)
			return
		}
	}

	siteName := os.Getenv(envWebAccelSiteName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			return nil
		},
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
