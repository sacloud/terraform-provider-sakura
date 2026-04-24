// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iam-api-go"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraIAMUserProvisioning_basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName1 := "sakura_iam_user_provisioning.foobar"
	rand := test.RandomName()

	var userProvisioning v1.ScimConfigurationBase
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraIAMUserProvisioningDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMUserProvisioning_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMUserProvisioningExists(resourceName1, &userProvisioning),
					resource.TestCheckResourceAttr(resourceName1, "name", rand),
					resource.TestCheckResourceAttr(resourceName1, "token_version", "1"),
					resource.TestCheckResourceAttrSet(resourceName1, "secret_token"),
					resource.TestCheckResourceAttrSet(resourceName1, "base_url"),
					resource.TestCheckResourceAttrSet(resourceName1, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName1, "updated_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMUserProvisioning_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMUserProvisioningExists(resourceName1, &userProvisioning),
					resource.TestCheckResourceAttr(resourceName1, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName1, "token_version", "2"),
					resource.TestCheckResourceAttrSet(resourceName1, "secret_token"),
					resource.TestCheckResourceAttrSet(resourceName1, "base_url"),
					resource.TestCheckResourceAttrSet(resourceName1, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName1, "updated_at"),
				),
			},
		},
	})
}

func testCheckSakuraIAMUserProvisioningExists(n string, up *v1.ScimConfigurationBase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no IAM Group ID is set")
		}

		client := test.AccClientGetter()
		scimOp := iam.NewScimOp(client.IamClient)
		foundUP, err := scimOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundUP.ID.String() != rs.Primary.ID {
			return fmt.Errorf("not found User Provisioning: %s", rs.Primary.ID)
		}

		*up = *foundUP
		return nil
	}
}

func testCheckSakuraIAMUserProvisioningDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	scimOp := iam.NewScimOp(client.IamClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_iam_user_provisioning" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := scimOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists IAM User Provisioning: %s", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraIAMUserProvisioning_basic = `
resource "sakura_iam_user_provisioning" "foobar" {
  name = "{{ .arg0 }}"
}`

const testAccSakuraIAMUserProvisioning_update = `
resource "sakura_iam_user_provisioning" "foobar" {
  name = "{{ .arg0 }}-upd"
  token_version = 2
}`
