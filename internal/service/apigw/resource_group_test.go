// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceAPIGWGroup_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION")

	resourceName := "sakura_apigw_group.foobar"
	rand := test.RandomName()
	var group v1.Group
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraAPIGWGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWGroup_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWGroup_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func testCheckSakuraAPIGWGroupDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	groupOp := apigw.NewGroupOp(client.ApigwClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apigw_group" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := groupOp.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists APIGW Group: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraAPIGWGroupExists(n string, group *v1.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no APIGW Group ID is set")
		}

		groupOp := apigw.NewGroupOp(test.AccClientGetter().ApigwClient)
		foundGroup, err := groupOp.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err != nil {
			return err
		}

		if foundGroup.ID.Value.String() != rs.Primary.ID {
			return fmt.Errorf("not found APIGW Group: %s", rs.Primary.ID)
		}

		*group = *foundGroup
		return nil
	}
}

var testAccSakuraAPIGWGroup_basic = testSetupAPIGWSub + `
resource "sakura_apigw_group" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1"]
}`

var testAccSakuraAPIGWGroup_update = testSetupAPIGWSub + `
resource "sakura_apigw_group" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1", "tag2"]
}`
