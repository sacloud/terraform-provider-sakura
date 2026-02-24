// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iam-api-go"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraIAMProject_basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "sakura_iam_project.foobar"
	rand := test.RandomName()

	var project v1.Project
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraIAMProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMProject_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "code", rand+"-code"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttrPair(resourceName, "parent_folder_id", "sakura_iam_folder.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMProject_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMProjectExists(resourceName, &project),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "code", rand+"-code"), // code cannot be updated
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
					resource.TestCheckResourceAttrPair(resourceName, "parent_folder_id", "sakura_iam_folder.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func testCheckSakuraIAMProjectExists(n string, project *v1.Project) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no IAM Project ID is set")
		}

		client := test.AccClientGetter()
		projectOp := iam.NewProjectOp(client.IamClient)
		foundProject, err := projectOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err != nil {
			return err
		}

		if strconv.Itoa(foundProject.ID) != rs.Primary.ID {
			return fmt.Errorf("not found Project: %s", rs.Primary.ID)
		}

		*project = *foundProject
		return nil
	}
}

func testCheckSakuraIAMProjectDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	projectOp := iam.NewProjectOp(client.IamClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_iam_project" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := projectOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err == nil {
			return fmt.Errorf("still exists IAM Project: %s", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraIAMProject_basic = `
resource "sakura_iam_folder" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_iam_project" "foobar" {
  name = "{{ .arg0 }}"
  code = "{{ .arg0 }}-code"
  description = "description"
  parent_folder_id = sakura_iam_folder.foobar.id
}
`

const testAccSakuraIAMProject_update = `
resource "sakura_iam_folder" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_iam_project" "foobar" {
  name = "{{ .arg0 }}-upd"
  code = "{{ .arg0 }}-code" # code cannot be updated
  description = "description-upd"
  parent_folder_id = sakura_iam_folder.foobar.id
}
`
