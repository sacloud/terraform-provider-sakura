// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package disk_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceDisk_basic(t *testing.T) {
	test.SkipIfFakeModeEnabled(t) // KMSを利用するためacctestでのみ実施したい
	test.SkipIfEnvIsNotSet(t, "SAKURA_KMS_KEY_ID")

	resourceName := "data.sakura_disk.foobar"
	rand := test.RandomName()
	kmsKeyID := os.Getenv("SAKURA_KMS_KEY_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceDisk_basic, rand, kmsKeyID),
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
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", kmsKeyID),
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
  kms_key_id           = "{{ .arg1 }}"
}

data "sakura_disk" "foobar" {
  name = sakura_disk.foobar.name
}`
