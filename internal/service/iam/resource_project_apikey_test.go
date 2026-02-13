// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iam-api-go"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraIAMProjectApiKey_basic(t *testing.T) {
	// 単にプロジェクトを作るだけでは権限が足りないので、あらかじめ十分な権限を持ったプロジェクトを指定する
	test.SkipIfEnvIsNotSet(t, "SAKURA_IAM_PROJECT_ID")

	resourceName := "sakura_iam_project_apikey.foobar"
	rand := test.RandomName()
	id := os.Getenv("SAKURA_IAM_PROJECT_ID")

	var apikey v1.ProjectApiKey
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraIAMProjectApiKeyDestroy,
			testCheckSakuraIAMProjectDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMProjectApiKey_basic, rand, id),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMProjectApiKeyExists(resourceName, &apikey),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.0", "resource-creator"),
					resource.TestCheckResourceAttrSet(resourceName, "access_token"),
					resource.TestCheckResourceAttrSet(resourceName, "access_token_secret"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckNoResourceAttr(resourceName, "zone"),
					resource.TestCheckNoResourceAttr(resourceName, "server_resource_id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMProjectApiKey_update, rand, id),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMProjectApiKeyExists(resourceName, &apikey),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.0", "resource-creator"),
					resource.TestCheckResourceAttr(resourceName, "iam_roles.1", "resource-viewer"),
					resource.TestCheckResourceAttrSet(resourceName, "access_token"),
					resource.TestCheckResourceAttrSet(resourceName, "access_token_secret"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckNoResourceAttr(resourceName, "zone"),
					resource.TestCheckNoResourceAttr(resourceName, "server_resource_id"),
				),
			},
		},
	})
}

func testCheckSakuraIAMProjectApiKeyExists(n string, apikey *v1.ProjectApiKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no IAM Project API Key ID is set")
		}

		client := test.AccClientGetter()
		projectApiKeyOp := iam.NewProjectAPIKeyOp(client.IamClient)
		foundProjectApiKey, err := projectApiKeyOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err != nil {
			return err
		}

		if strconv.Itoa(foundProjectApiKey.ID) != rs.Primary.ID {
			return fmt.Errorf("not found IAM Project API Key: %s", rs.Primary.ID)
		}

		*apikey = *foundProjectApiKey
		return nil
	}
}

func testCheckSakuraIAMProjectApiKeyDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	projectApiKeyOp := iam.NewProjectAPIKeyOp(client.IamClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_iam_project_apikey" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := projectApiKeyOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err == nil {
			return fmt.Errorf("still exists IAM Project API Key: %s", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraIAMProjectApiKey_basic = `
resource "sakura_iam_project_apikey" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  project_id = "{{ .arg1 }}"
  iam_roles = ["resource-creator"]
}
`

const testAccSakuraIAMProjectApiKey_update = `
resource "sakura_iam_project_apikey" "foobar" {
  name = "{{ .arg0 }}-upd"
  description = "description-upd"
  project_id = "{{ .arg1 }}"
  iam_roles = ["resource-creator", "resource-viewer"]
}
`
