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

func TestAccSakuraIAMServicePrincipal_basic(t *testing.T) {
	resourceName := "sakura_iam_service_principal.foobar"
	rand := test.RandomName()

	var servicePrincipal v1.ServicePrincipal
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraIAMServicePrincipalDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMServicePrincipal_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMServicePrincipalExists(resourceName, &servicePrincipal),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", "sakura_iam_project.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraIAMServicePrincipal_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIAMServicePrincipalExists(resourceName, &servicePrincipal),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
					resource.TestCheckResourceAttrPair(resourceName, "project_id", "sakura_iam_project.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
		},
	})
}

func testCheckSakuraIAMServicePrincipalExists(n string, servicePrincipal *v1.ServicePrincipal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no IAM Service Principal ID is set")
		}

		client := test.AccClientGetter()
		servicePrincipalOp := iam.NewServicePrincipalOp(client.IamClient)
		foundServicePrincipal, err := servicePrincipalOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err != nil {
			return err
		}

		if strconv.Itoa(foundServicePrincipal.ID) != rs.Primary.ID {
			return fmt.Errorf("not found IAM Service Principal: %s", rs.Primary.ID)
		}

		*servicePrincipal = *foundServicePrincipal
		return nil
	}
}

func testCheckSakuraIAMServicePrincipalDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	servicePrincipalOp := iam.NewServicePrincipalOp(client.IamClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_iam_service_principal" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := servicePrincipalOp.Read(context.Background(), utils.MustAtoI((rs.Primary.ID)))
		if err == nil {
			return fmt.Errorf("still exists IAM Service Principal: %s", rs.Primary.ID)
		}
	}

	return nil
}

const testAccSakuraIAMServicePrincipal_basic = `
resource "sakura_iam_project" "foobar" {
  name = "{{ .arg0 }}"
  code = "{{ .arg0 }}-code"
  description = "description"
}

resource "sakura_iam_service_principal" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  project_id  = sakura_iam_project.foobar.id
}
`

const testAccSakuraIAMServicePrincipal_update = `
resource "sakura_iam_project" "foobar" {
  name = "{{ .arg0 }}"
  code = "{{ .arg0 }}-code"
  description = "description"
}

resource "sakura_iam_service_principal" "foobar" {
  name = "{{ .arg0 }}-upd"
  description = "description-upd"
  project_id  = sakura_iam_project.foobar.id
}
`
