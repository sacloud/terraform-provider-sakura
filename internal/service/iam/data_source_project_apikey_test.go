// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceIAMProjectApiKey_Basic(t *testing.T) {
	// 単にプロジェクトを作るだけでは権限が足りないので、あらかじめ十分な権限を持ったプロジェクトを指定する
	test.SkipIfEnvIsNotSet(t, "SAKURA_IAM_PROJECT_ID")

	resourceName := "data.sakura_iam_project_apikey.foobar"
	rand := test.RandomName()
	id := os.Getenv("SAKURA_IAM_PROJECT_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraDataSourceIAMProjectApiKeyConfig, rand, id),
				Check: resource.ComposeTestCheckFunc(
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.0", "resource-creator"),
					resource.TestCheckResourceAttrSet(resourceName, "access_token"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckNoResourceAttr(resourceName, "zone"),
					resource.TestCheckNoResourceAttr(resourceName, "server_resource_id"),
				),
			},
		},
	})
}

//nolint:gosec
const testAccCheckSakuraDataSourceIAMProjectApiKeyConfig = `
resource "sakura_iam_project_apikey" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  project_id = "{{ .arg1 }}"
  iam_roles = ["resource-creator"]
}

data "sakura_iam_project_apikey" "foobar" {
  name = sakura_iam_project_apikey.foobar.name
}`
