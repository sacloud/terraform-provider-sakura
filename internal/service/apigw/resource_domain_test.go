// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceAPIGWDomain_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION")

	resourceName := "sakura_apigw_domain.foobar"
	rand := test.RandomName()
	domainName := fmt.Sprintf("%s.com", rand)
	var domain v1.Domain
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraAPIGWDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWDomain_basic, domainName),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
					resource.TestCheckResourceAttr(resourceName, "certificate_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWDomain_update, domainName, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
					resource.TestCheckResourceAttr(resourceName, "certificate_name", rand),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_id", "sakura_apigw_cert.foobar", "id"),
				),
			},
		},
	})
}

func testCheckSakuraAPIGWDomainDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apigw_domain" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := getAPIGWDomainByID(client.ApigwClient, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists APIGW Domain: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraAPIGWDomainExists(n string, domain *v1.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no APIGW Domain ID is set")
		}

		client := test.AccClientGetter()
		foundDomain, err := getAPIGWDomainByID(client.ApigwClient, rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundDomain.ID.Value.String() != rs.Primary.ID {
			return fmt.Errorf("not found APIGW Domain: %s", rs.Primary.ID)
		}

		*domain = *foundDomain
		return nil
	}
}

func getAPIGWDomainByID(client *v1.Client, id string) (*v1.Domain, error) {
	domainOp := apigw.NewDomainOp(client)
	domains, err := domainOp.List(context.Background())
	if err != nil {
		return nil, err
	}

	var domain *v1.Domain
	for _, d := range domains {
		if id != "" && d.ID.Value.String() == id {
			domain = &d
			break
		}
	}
	if domain == nil {
		return nil, fmt.Errorf("APIGW Domain not found: %s", id)
	}

	return domain, nil
}

var testAccSakuraAPIGWDomain_basic = testSetupAPIGWSub + `
resource "sakura_apigw_domain" "foobar" {
  name = "{{ .arg0 }}"
}`

var testAccSakuraAPIGWDomain_update = testSetupAPIGWSub + `
resource "sakura_apigw_cert" "foobar" {
  name = "{{ .arg1 }}"
  rsa  = {
    cert_wo = file("testdata/rsa.crt")
    key_wo = file("testdata/rsa.key")
    cert_wo_version = 1
  }
}

resource "sakura_apigw_domain" "foobar" {
  name = "{{ .arg0 }}"
  certificate_id = sakura_apigw_cert.foobar.id
}`
