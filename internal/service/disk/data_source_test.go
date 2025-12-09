// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package disk_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceDisk_basic(t *testing.T) {
	test.SkipIfFakeModeEnabled(t) // KMSを利用するためacctestでのみ実施したい

	resourceName := "data.sakura_disk.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceDisk_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "plan", "ssd"),
					resource.TestCheckResourceAttr(resourceName, "connector", "virtio"),
					resource.TestCheckResourceAttr(resourceName, "size", "20"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
					resource.TestCheckResourceAttr(resourceName, "encryption_algorithm", "aes256_xts"),
					resource.TestCheckResourceAttrPair(
						resourceName, "kms_key_id",
						"sakura_kms.foobar", "id",
					),
				),
			},
		},
	})
}

var testAccSakuraDataSourceDisk_basic = `
resource "sakura_disk" "foobar"{
  name        = "{{ .arg0 }}"
  tags        = ["tag1", "tag2", "tag3"]
  description = "description"

  encryption_algorithm = "aes256_xts"
  kms_key_id           = sakura_kms.foobar.id
}

resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
}

data "sakura_disk" "foobar" {
  name = sakura_disk.foobar.name
}`
