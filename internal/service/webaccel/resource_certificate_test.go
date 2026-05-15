// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package webaccel_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
	"github.com/sacloud/webaccel-api-go"
)

const (
	envWebAccelCertificateCrt    = "SAKURA_WEBACCEL_CERT_PATH"
	envWebAccelCertificateKey    = "SAKURA_WEBACCEL_KEY_PATH"
	envWebAccelCertificateCrtUpd = "SAKURA_WEBACCEL_CERT_PATH_UPD"
	envWebAccelCertificateKeyUpd = "SAKURA_WEBACCEL_KEY_PATH_UPD"
)

func TestAccResourceSakuraWebAccelCertificate_basic(t *testing.T) {
	envKeys := []string{
		envWebAccelSiteName,
		envWebAccelCertificateCrt,
		envWebAccelCertificateKey,
		envWebAccelCertificateCrtUpd,
		envWebAccelCertificateKeyUpd,
	}
	for _, k := range envKeys {
		if os.Getenv(k) == "" {
			t.Skipf("ENV %q is required. skip", k)
			return
		}
	}

	siteName := os.Getenv(envWebAccelSiteName)
	crt := os.Getenv(envWebAccelCertificateCrt)
	key := os.Getenv(envWebAccelCertificateKey)
	crtUpd := os.Getenv(envWebAccelCertificateCrtUpd)
	keyUpd := os.Getenv(envWebAccelCertificateKeyUpd)

	regexpNotEmpty := regexp.MustCompile(".+")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraWebAccelCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraWebAccelCertificateConfig(siteName, crt, key, "1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "id", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "site_id", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "not_before", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "not_after", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "issuer_common_name", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "subject_common_name", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "sha256_fingerprint", regexpNotEmpty),
					resource.TestCheckResourceAttr("sakura_webaccel_certificate.foobar", "certificate_wo_version", "1"),
					resource.TestCheckNoResourceAttr("sakura_webaccel_certificate.foobar", "certificate_chain_wo"),
					resource.TestCheckNoResourceAttr("sakura_webaccel_certificate.foobar", "private_key_wo"),
				),
			},
			{
				Config: testAccCheckSakuraWebAccelCertificateConfig(siteName, crtUpd, keyUpd, "2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "id", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "site_id", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "not_before", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "not_after", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "issuer_common_name", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "subject_common_name", regexpNotEmpty),
					resource.TestMatchResourceAttr("sakura_webaccel_certificate.foobar", "sha256_fingerprint", regexpNotEmpty),
					resource.TestCheckResourceAttr("sakura_webaccel_certificate.foobar", "certificate_wo_version", "2"),
					resource.TestCheckNoResourceAttr("sakura_webaccel_certificate.foobar", "certificate_chain_wo"),
					resource.TestCheckNoResourceAttr("sakura_webaccel_certificate.foobar", "private_key_wo"),
				),
			},
		},
	})
}

func testCheckSakuraWebAccelCertificateDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := webaccel.NewOp(client.WebaccelClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_webaccel_certificate" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		res, err := op.ReadCertificate(context.Background(), rs.Primary.ID)
		if err == nil && res.Current != nil {
			return fmt.Errorf("still exists WebAccel Certificate: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckSakuraWebAccelCertificateConfig(name, crt, key, version string) string {
	tmpl := `
data "sakura_webaccel" "site" {
  name = "%s"
}
resource "sakura_webaccel_certificate" "foobar" {
  site_id           = data.sakura_webaccel.site.id
  certificate_chain_wo = file("%s")
  private_key_wo       = file("%s")
  certificate_wo_version = %s
}
`
	return fmt.Sprintf(tmpl, name, crt, key, version)
}
