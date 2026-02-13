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

func TestAccSakuraIAMSSO_basic(t *testing.T) {
	resourceName := "sakura_iam_sso.foobar"
	rand := test.RandomName()

	var sso v1.SSOProfile
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraIAMProjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMSSO_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMSSOExists(resourceName, &sso),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "idp_entity_id", "https://idp.example.com/ile2ephei7saeph6"),
					resource.TestCheckResourceAttr(resourceName, "idp_login_url", "https://idp.example.com/ile2ephei7saeph6/sso/login"),
					resource.TestCheckResourceAttr(resourceName, "idp_logout_url", "https://idp.example.com/ile2ephei7saeph6/sso/logout"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "sp_entity_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sp_acs_url"),
					resource.TestCheckResourceAttrSet(resourceName, "assigned"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMSSO_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMSSOExists(resourceName, &sso),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
					resource.TestCheckResourceAttr(resourceName, "idp_entity_id", "https://idp.example.com/ile2ephei7saeph6-2"),
					resource.TestCheckResourceAttr(resourceName, "idp_login_url", "https://idp.example.com/ile2ephei7saeph6-2/sso/login"),
					resource.TestCheckResourceAttr(resourceName, "idp_logout_url", "https://idp.example.com/ile2ephei7saeph6-2/sso/logout"),
					resource.TestCheckResourceAttrSet(resourceName, "idp_certificate"),
					resource.TestCheckResourceAttrSet(resourceName, "sp_entity_id"),
					resource.TestCheckResourceAttrSet(resourceName, "sp_acs_url"),
					resource.TestCheckResourceAttrSet(resourceName, "assigned"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func testCheckSakuraIAMSSOExists(n string, sso *v1.SSOProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no IAM SSO ID is set")
		}

		client := test.AccClientGetter()
		ssoOp := iam.NewSSOOp(client.IamClient)
		foundSSO, err := ssoOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err != nil {
			return err
		}

		if strconv.Itoa(foundSSO.ID) != rs.Primary.ID {
			return fmt.Errorf("not found SSO: %s", rs.Primary.ID)
		}

		*sso = *foundSSO
		return nil
	}
}

func testCheckSakuraIAMSSODestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	ssoOp := iam.NewSSOOp(client.IamClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_iam_sso" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := ssoOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err == nil {
			return fmt.Errorf("still exists IAM SSO: %s", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraIAMSSO_basic = `
resource "sakura_iam_sso" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  idp_entity_id = "https://idp.example.com/ile2ephei7saeph6"
  idp_login_url = "https://idp.example.com/ile2ephei7saeph6/sso/login"
  idp_logout_url = "https://idp.example.com/ile2ephei7saeph6/sso/logout"
  idp_certificate = file("testdata/rsa.crt")
}
`

const testAccSakuraIAMSSO_update = `
resource "sakura_iam_sso" "foobar" {
  name = "{{ .arg0 }}-upd"
  description = "description-upd"
  idp_entity_id = "https://idp.example.com/ile2ephei7saeph6-2"
  idp_login_url = "https://idp.example.com/ile2ephei7saeph6-2/sso/login"
  idp_logout_url = "https://idp.example.com/ile2ephei7saeph6-2/sso/logout"
  idp_certificate = file("testdata/rsa.crt")
}
`
