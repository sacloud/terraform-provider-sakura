// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package webaccel_test

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

const envWebAccelDomain = "SAKURA_WEBACCEL_DOMAIN"

func TestAccSakuraDataSourceWebAccel_ByName(t *testing.T) {
	var siteName string
	if name, ok := os.LookupEnv(envWebAccelSiteName); ok {
		siteName = name
	} else {
		t.Skipf("ENV %q is required. skip", envWebAccelSiteName)
		return
	}

	regexpNotEmpty := regexp.MustCompile(".+")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraCloudDataSourceWebAccelWithName(siteName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSakuraCloudWebAccelDataSourceID("data.sakura_webaccel.foobar"),
					resource.TestCheckResourceAttr("data.sakura_webaccel.foobar", "name", siteName),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "domain", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "origin", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "request_protocol", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "origin_parameters.type", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "origin_parameters.origin", regexpNotEmpty),
					// resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "logging.0.enabled", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "subdomain", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "domain_type", regexpNotEmpty),
					// resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "has_certificate", regexpNotEmpty),
					// resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "host_header", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "status", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "cname_record_value", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "txt_record_value", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "vary_support", regexpNotEmpty),
				),
			},
		},
	})
}

func TestAccSakuraDataSourceWebAccel_ByDomain(t *testing.T) {
	var domainName string
	if name, ok := os.LookupEnv(envWebAccelDomain); ok {
		domainName = name
	} else {
		t.Skipf("ENV %q is required. skip", envWebAccelDomain)
		return
	}

	regexpNotEmpty := regexp.MustCompile(".+")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSakuraCloudDataSourceWebAccelWithDomain(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSakuraCloudWebAccelDataSourceID("data.sakura_webaccel.foobar"),
					resource.TestCheckResourceAttr("data.sakura_webaccel.foobar", "domain", domainName),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "name", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "origin", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "origin_parameters.type", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "subdomain", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "domain_type", regexpNotEmpty),
					// resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "has_certificate", regexpNotEmpty),
					// resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "host_header", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "status", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "cname_record_value", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "txt_record_value", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "vary_support", regexpNotEmpty),
					resource.TestMatchResourceAttr("data.sakura_webaccel.foobar", "normalize_ae", regexpNotEmpty),
				),
			},
		},
	})
}

func testAccCheckSakuraCloudWebAccelDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: source: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("WebAccel data source ID not set")
		}
		return nil
	}
}

func testAccCheckSakuraCloudDataSourceWebAccelWithName(name string) string {
	tmpl := `
data "sakura_webaccel" "foobar" {
  name = "%s"
}`
	return fmt.Sprintf(tmpl, name)
}

func testAccCheckSakuraCloudDataSourceWebAccelWithDomain(domain string) string {
	tmpl := `
data "sakura_webaccel" "foobar" {
  domain = "%s"
}`
	return fmt.Sprintf(tmpl, domain)
}
