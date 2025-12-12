// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package dns_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDNSDataSource_basic(t *testing.T) {
	resourceName := "data.sakura_dns.foobar"
	zone := fmt.Sprintf("%s.com", test.RandomName())

	var dns iaas.DNS
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceDNS_basic, zone),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraDNSExists(resourceName, &dns),
					resource.TestCheckResourceAttr(resourceName, "zone", zone),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.2", "tag3"),
					resource.TestCheckResourceAttr(resourceName, "record.0.name", "www"),
					resource.TestCheckResourceAttr(resourceName, "record.0.type", "A"),
					resource.TestCheckResourceAttr(resourceName, "record.0.value", "192.168.11.1"),
					resource.TestCheckResourceAttr(resourceName, "record.1.name", "www2"),
					resource.TestCheckResourceAttr(resourceName, "record.1.type", "A"),
					resource.TestCheckResourceAttr(resourceName, "record.1.value", "192.168.11.2"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceDNS_basic = `
resource "sakura_dns" "foobar" {
  zone = "{{ .arg0 }}"
  description = "description"
  tags = ["tag1", "tag2", "tag3"]
}
resource "sakura_dns_record" "foobar1" {
  dns_id = sakura_dns.foobar.id
  name  = "www"
  type  = "A"
  value = "192.168.11.1"
}
resource "sakura_dns_record" "foobar2" {
  dns_id = sakura_dns.foobar.id
  name  = "www2"
  type  = "A"
  value = "192.168.11.2"

  depends_on = [sakura_dns_record.foobar1] # to ensure creation order
}

data "sakura_dns" "foobar" {
  zone = sakura_dns.foobar.zone

  depends_on = [sakura_dns_record.foobar2]
}`
