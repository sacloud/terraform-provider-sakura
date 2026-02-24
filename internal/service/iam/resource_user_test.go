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

func TestAccSakuraIAMUser_basic(t *testing.T) {
	test.SkipIfIAMEnvIsNotSet(t)

	resourceName := "sakura_iam_user.foobar"
	rand := test.RandomName()

	var user v1.User
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraIAMUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMUser_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "code", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "email", "tf-test-email@sakura.ad.jp"),
					resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
					resource.TestCheckResourceAttr(resourceName, "password_wo_version", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "passwordless"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "otp.status"),
					resource.TestCheckResourceAttrSet(resourceName, "otp.has_recovery_code"),
					resource.TestCheckResourceAttrSet(resourceName, "member.id"),
					resource.TestCheckResourceAttrSet(resourceName, "member.code"),
					resource.TestCheckResourceAttrSet(resourceName, "security_key_registered"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMUser_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "code", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
					resource.TestCheckResourceAttr(resourceName, "email", "tf-test-email@sakura.ad.jp"),
					resource.TestCheckNoResourceAttr(resourceName, "password_wo"),
					resource.TestCheckResourceAttr(resourceName, "password_wo_version", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "passwordless"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "otp.status"),
					resource.TestCheckResourceAttrSet(resourceName, "otp.has_recovery_code"),
					resource.TestCheckResourceAttrSet(resourceName, "member.id"),
					resource.TestCheckResourceAttrSet(resourceName, "member.code"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func testCheckSakuraIAMUserExists(n string, user *v1.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no IAM User ID is set")
		}

		client := test.AccClientGetter()
		userOp := iam.NewUserOp(client.IamClient)
		foundUser, err := userOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err != nil {
			return err
		}

		if strconv.Itoa(foundUser.ID) != rs.Primary.ID {
			return fmt.Errorf("not found IAM User: %s", rs.Primary.ID)
		}

		*user = *foundUser
		return nil
	}
}

func testCheckSakuraIAMUserDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	userOp := iam.NewUserOp(client.IamClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_iam_user" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := userOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err == nil {
			return fmt.Errorf("still exists IAM User: %s", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraIAMUser_basic = `
resource "sakura_iam_user" "foobar" {
  name = "{{ .arg0 }}"
  code = "{{ .arg0 }}"
  description = "description"
  email = "tf-test-email@sakura.ad.jp"
  password_wo = "PassWord!1234567"
  password_wo_version = 1
}
`

const testAccSakuraIAMUser_update = `
resource "sakura_iam_user" "foobar" {
  name = "{{ .arg0 }}-upd"
  code = "{{ .arg0 }}"
  description = "description-upd"
  email = "tf-test-email@sakura.ad.jp"
  password_wo = "PassWord!1234567"
  password_wo_version = 1
}
`
