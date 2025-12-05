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

data "sakura_dns" "foobar" {
  name = sakura_dns.foobar.zone
}`
