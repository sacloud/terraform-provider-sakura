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

func TestAccSakuraIAMFolder_basic(t *testing.T) {
	resourceName1 := "sakura_iam_folder.foobar1"
	resourceName2 := "sakura_iam_folder.foobar2"
	rand := test.RandomName()
	name1 := rand + "-1"
	name2 := rand + "-2"

	var folder v1.Folder
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraIAMFolderDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMFolder_basic, name1, name2),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMFolderExists(resourceName1, &folder),
					resource.TestCheckResourceAttr(resourceName1, "name", name1),
					resource.TestCheckResourceAttr(resourceName1, "description", "description1"),
					resource.TestCheckNoResourceAttr(resourceName1, "parent_id"),
					resource.TestCheckResourceAttrSet(resourceName1, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName1, "updated_at"),
					testCheckSakuraIAMFolderExists(resourceName2, &folder),
					resource.TestCheckResourceAttr(resourceName2, "name", name2),
					resource.TestCheckResourceAttr(resourceName2, "description", "description2"),
					resource.TestCheckResourceAttrPair(resourceName2, "parent_id", "sakura_iam_folder.foobar1", "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName2, "updated_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMFolder_update, name1, name2),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMFolderExists(resourceName1, &folder),
					resource.TestCheckResourceAttr(resourceName1, "name", name1+"-upd"),
					resource.TestCheckResourceAttr(resourceName1, "description", "description1-upd"),
					resource.TestCheckNoResourceAttr(resourceName1, "parent_id"),
					resource.TestCheckResourceAttrSet(resourceName1, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName1, "updated_at"),
					testCheckSakuraIAMFolderExists(resourceName2, &folder),
					resource.TestCheckResourceAttr(resourceName2, "name", name2+"-upd"),
					resource.TestCheckResourceAttr(resourceName2, "description", "description2-upd"),
					resource.TestCheckResourceAttrPair(resourceName2, "parent_id", "sakura_iam_folder.foobar1", "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName2, "updated_at"),
				),
			},
		},
	})
}

func testCheckSakuraIAMFolderExists(n string, folder *v1.Folder) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no IAM Folder ID is set")
		}

		client := test.AccClientGetter()
		folderOp := iam.NewFolderOp(client.IamClient)
		foundFolder, err := folderOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err != nil {
			return err
		}

		if strconv.Itoa(foundFolder.ID) != rs.Primary.ID {
			return fmt.Errorf("not found Folder: %s", rs.Primary.ID)
		}

		*folder = *foundFolder
		return nil
	}
}

func testCheckSakuraIAMFolderDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	folderOp := iam.NewFolderOp(client.IamClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_iam_folder" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := folderOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err == nil {
			return fmt.Errorf("still exists IAM Folder: %s", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraIAMFolder_basic = `
resource "sakura_iam_folder" "foobar1" {
  name = "{{ .arg0 }}"
  description = "description1"
}

resource "sakura_iam_folder" "foobar2" {
  name = "{{ .arg1 }}"
  description = "description2"
  parent_id = sakura_iam_folder.foobar1.id
}
`

const testAccSakuraIAMFolder_update = `
resource "sakura_iam_folder" "foobar1" {
  name = "{{ .arg0 }}-upd"
  description = "description1-upd"
}

resource "sakura_iam_folder" "foobar2" {
  name = "{{ .arg1 }}-upd"
  description = "description2-upd"
  parent_id = sakura_iam_folder.foobar1.id
}`
