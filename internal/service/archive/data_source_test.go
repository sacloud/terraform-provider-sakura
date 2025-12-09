// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package archive_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceArchive_osType(t *testing.T) {
	resourceName := "data.sakura_archive.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraDataSourceArchive_osType,
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
				),
			},
		},
	})
}

func TestAccSakuraDataSourceArchive_withTag(t *testing.T) {
	resourceName := "data.sakura_archive.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraDataSourceArchive_withTag,
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceArchive_withTag = `
data "sakura_archive" "foobar" {
    tags = ["distro-ubuntu","os-linux"]
}`

var testAccCheckSakuraDataSourceArchive_osType = `
data "sakura_archive" "foobar" {
    os_type = "ubuntu"
}
`
