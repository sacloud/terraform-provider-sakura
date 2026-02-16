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

func TestAccSakuraIAMGroup_basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName1 := "sakura_iam_group.foobar"
	rand := test.RandomName()

	var group v1.Group
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraIAMGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMGroup_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMGroupExists(resourceName1, &group),
					resource.TestCheckResourceAttr(resourceName1, "name", rand),
					resource.TestCheckResourceAttr(resourceName1, "description", "description1"),
					resource.TestCheckResourceAttrSet(resourceName1, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName1, "updated_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMGroup_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMGroupExists(resourceName1, &group),
					resource.TestCheckResourceAttr(resourceName1, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName1, "description", "description1-upd"),
					resource.TestCheckResourceAttrSet(resourceName1, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName1, "updated_at"),
				),
			},
		},
	})
}

func testCheckSakuraIAMGroupExists(n string, group *v1.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no IAM Group ID is set")
		}

		client := test.AccClientGetter()
		groupOp := iam.NewGroupOp(client.IamClient)
		foundGroup, err := groupOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err != nil {
			return err
		}

		if strconv.Itoa(foundGroup.ID) != rs.Primary.ID {
			return fmt.Errorf("not found Group: %s", rs.Primary.ID)
		}

		*group = *foundGroup
		return nil
	}
}

func testCheckSakuraIAMGroupDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	groupOp := iam.NewGroupOp(client.IamClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_iam_group" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := groupOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err == nil {
			return fmt.Errorf("still exists IAM Group: %s", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraIAMGroup_basic = `
resource "sakura_iam_group" "foobar" {
  name = "{{ .arg0 }}"
  description = "description1"
}
`

const testAccSakuraIAMGroup_update = `
resource "sakura_iam_group" "foobar" {
  name = "{{ .arg0 }}-upd"
  description = "description1-upd"
}
`
