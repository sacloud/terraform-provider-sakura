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

func TestAccSakuraResourceAPIGWCert_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION")

	resourceName := "sakura_apigw_cert.foobar"
	rand := test.RandomName()
	var cert v1.Certificate
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraAPIGWCertDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWCert_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWCertExists(resourceName, &cert),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "rsa.cert_wo_version", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "rsa.cert_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "rsa.key_wo"),
					resource.TestCheckResourceAttrSet(resourceName, "rsa.expired_at"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWCert_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWCertExists(resourceName, &cert),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "rsa.cert_wo_version", "2"),
					resource.TestCheckNoResourceAttr(resourceName, "rsa.cert_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "rsa.key_wo"),
					resource.TestCheckResourceAttrSet(resourceName, "rsa.expired_at"),
					resource.TestCheckResourceAttr(resourceName, "ecdsa.cert_wo_version", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "ecdsa.cert_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "ecdsa.key_wo"),
					resource.TestCheckResourceAttrSet(resourceName, "ecdsa.expired_at"),
				),
			},
		},
	})
}

func testCheckSakuraAPIGWCertDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apigw_cert" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := getAPIGWCertByID(client.ApigwClient, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists APIGW Cert: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraAPIGWCertExists(n string, cert *v1.Certificate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no APIGW Cert ID is set")
		}

		client := test.AccClientGetter()
		foundCert, err := getAPIGWCertByID(client.ApigwClient, rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundCert.ID.Value.String() != rs.Primary.ID {
			return fmt.Errorf("not found APIGW Cert: %s", rs.Primary.ID)
		}

		*cert = *foundCert
		return nil
	}
}

func getAPIGWCertByID(client *v1.Client, id string) (*v1.Certificate, error) {
	certOp := apigw.NewCertificateOp(client)
	certs, err := certOp.List(context.Background())
	if err != nil {
		return nil, err
	}

	var cert *v1.Certificate
	for _, c := range certs {
		if id != "" && c.ID.Value.String() == id {
			cert = &c
			break
		}
	}
	if cert == nil {
		return nil, fmt.Errorf("APIGW Cert not found: %s", id)
	}

	return cert, nil
}

var testAccSakuraAPIGWCert_basic = testSetupAPIGWSub + `
resource "sakura_apigw_cert" "foobar" {
  name = "{{ .arg0 }}"
  rsa  = {
    cert_wo = file("testdata/rsa.crt")
    key_wo = file("testdata/rsa.key")
    cert_wo_version = 1
  }
}`

var testAccSakuraAPIGWCert_update = testSetupAPIGWSub + `
resource "sakura_apigw_cert" "foobar" {
  name = "{{ .arg0 }}"
  rsa  = {
    cert_wo = file("testdata/rsa.crt")
    key_wo = file("testdata/rsa.key")
    cert_wo_version = 2
  }
  ecdsa = {
    cert_wo = file("testdata/ecdsa.crt")
    key_wo = file("testdata/ecdsa.key")
    cert_wo_version = 1
  }
}`
