// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceIAMProject_Basic(t *testing.T) {
	resourceName := "data.sakura_iam_project.foobar"
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceIAMProjectConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "code", rand+"-code"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckNoResourceAttr(resourceName, "parent_folder_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

var testAccCheckSakuraDataSourceIAMProjectConfig = `
resource "sakura_iam_project" "foobar" {
  name        = "{{ .arg0 }}"
  code        = "{{ .arg0 }}-code"
  description = "description"
}

data "sakura_iam_project" "foobar" {
  name = sakura_iam_project.foobar.name
}`
