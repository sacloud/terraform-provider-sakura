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

func TestAccSakuraResourceAPIGWUser_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION")

	resourceName := "sakura_apigw_user.foobar"
	rand := test.RandomName()
	var user v1.UserDetail
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraAPIGWUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWUser_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "custom_id", "custom-test-id-1000"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.protocols", "http"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.restricted_by", "allowIps"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.0", "192.168.0.10"),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", rand),
					resource.TestCheckResourceAttr(resourceName, "authentication.basic_auth.username", rand+"-user"),
					resource.TestCheckNoResourceAttr(resourceName, "authentication.basic_auth.password_wo"),
					resource.TestCheckResourceAttr(resourceName, "authentication.basic_auth.password_wo_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication.jwt.key", rand+"-key"),
					resource.TestCheckNoResourceAttr(resourceName, "authentication.jwt.secret_wo"),
					resource.TestCheckResourceAttr(resourceName, "authentication.jwt.secret_wo_version", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "authentication.hmac_auth"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWUser_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "custom_id", "custom-test-id-2000"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.protocols", "http,https"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.restricted_by", "denyIps"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.0", "192.168.0.1"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.1", "192.168.0.2"),
					resource.TestCheckNoResourceAttr(resourceName, "groups"),
					resource.TestCheckNoResourceAttr(resourceName, "authentication.basic_auth"),
					resource.TestCheckResourceAttr(resourceName, "authentication.jwt.key", rand+"-key2"),
					resource.TestCheckNoResourceAttr(resourceName, "authentication.jwt.secret_wo"),
					resource.TestCheckResourceAttr(resourceName, "authentication.jwt.secret_wo_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "authentication.hmac_auth.username", rand+"-auth"),
					resource.TestCheckNoResourceAttr(resourceName, "authentication.hmac_auth.secret_wo"),
					resource.TestCheckResourceAttr(resourceName, "authentication.hmac_auth.secret_wo_version", "1"),
				),
			},
		},
	})
}

func testCheckSakuraAPIGWUserDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	userOp := apigw.NewUserOp(client.ApigwClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apigw_user" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := userOp.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists APIGW User: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraAPIGWUserExists(n string, user *v1.UserDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no APIGW User ID is set")
		}

		userOp := apigw.NewUserOp(test.AccClientGetter().ApigwClient)
		foundUser, err := userOp.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err != nil {
			return err
		}

		if foundUser.ID.Value.String() != rs.Primary.ID {
			return fmt.Errorf("not found APIGW User: %s", rs.Primary.ID)
		}

		*user = *foundUser
		return nil
	}
}

var testAccSakuraAPIGWUser_basic = testSetupAPIGWSub + `
resource "sakura_apigw_group" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_apigw_user" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1"]
  custom_id = "custom-test-id-1000"
  ip_restriction = {
    protocols = "http"
    restricted_by = "allowIps"
    ips = ["192.168.0.10"]
  }
  groups = [{name = "{{ .arg0 }}"}]
  authentication = {
    basic_auth = {
       username = "{{ .arg0 }}-user",
       password_wo = "password"
	   password_wo_version = 1
    },
    jwt = {
      key = "{{ .arg0 }}-key",
      secret_wo = "secret",
	  secret_wo_version = 1,
      algorithm = "HS256"
    },
  }
}`

var testAccSakuraAPIGWUser_update = testSetupAPIGWSub + `
resource "sakura_apigw_group" "foobar" {
  name = "{{ .arg0 }}"
}

resource "sakura_apigw_user" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1", "tag2"]
  custom_id = "custom-test-id-2000"
  ip_restriction = {
    protocols = "http,https"
    restricted_by = "denyIps"
    ips = ["192.168.0.1", "192.168.0.2"]
  }
  authentication = {
    jwt = {
      key = "{{ .arg0 }}-key2",
      secret_wo = "secret",
	  secret_wo_version = 2,
      algorithm = "HS256"
    },
	hmac_auth = {
	  username = "{{ .arg0 }}-auth",
	  secret_wo = "secret",
	  secret_wo_version = 1
	}
  }
}`
