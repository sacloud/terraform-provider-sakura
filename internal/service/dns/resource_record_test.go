// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package dns_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/service/dns"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDNSRecord_basic(t *testing.T) {
	resourceName1 := "sakura_dns_record.foobar1"
	resourceName2 := "sakura_dns_record.foobar2"

	zone := fmt.Sprintf("%s.com", test.RandomName())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			test.CheckSakuraDNSDestroy,
			testCheckSakuraDNSRecordDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDNSRecord_basic, zone),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "name", "www"),
					resource.TestCheckResourceAttr(resourceName1, "type", "A"),
					resource.TestCheckResourceAttr(resourceName1, "value", "192.168.0.1"),
					resource.TestCheckResourceAttr(resourceName2, "name", "_sip._tls"),
					resource.TestCheckResourceAttr(resourceName2, "type", "SRV"),
					resource.TestCheckResourceAttr(resourceName2, "value", "www.sakura.ne.jp."),
					resource.TestCheckResourceAttr(resourceName2, "priority", "1"),
					resource.TestCheckResourceAttr(resourceName2, "weight", "2"),
					resource.TestCheckResourceAttr(resourceName2, "port", "3"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDNSRecord_update, zone),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "name", "www2"),
					resource.TestCheckResourceAttr(resourceName1, "type", "A"),
					resource.TestCheckResourceAttr(resourceName1, "value", "192.168.0.2"),
				),
			},
		},
	})
}

func TestAccSakuraDNSRecord_withCount(t *testing.T) {
	resourceName := "sakura_dns_record.foobar"
	zone := fmt.Sprintf("%s.com", test.RandomName())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			test.CheckSakuraDNSDestroy,
			testCheckSakuraDNSRecordDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDNSRecord_withCount, zone),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName+".0", "name", "www"),
					resource.TestCheckResourceAttr(resourceName+".0", "type", "A"),
					resource.TestCheckResourceAttr(resourceName+".0", "value", "192.168.0.1"),
					resource.TestCheckResourceAttr(resourceName+".1", "name", "www"),
					resource.TestCheckResourceAttr(resourceName+".1", "type", "A"),
					resource.TestCheckResourceAttr(resourceName+".1", "value", "192.168.0.2"),
				),
			},
		},
	})
}

func TestAccSakuraDNSRecord_ImportWithIdentity(t *testing.T) {
	resourceName := "sakura_dns_record.foobar1"
	zone := fmt.Sprintf("%s.com", test.RandomName())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			test.CheckSakuraDNSDestroy,
			testCheckSakuraDNSRecordDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDNSRecord_update, zone),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "www2"),
					resource.TestCheckResourceAttr(resourceName, "type", "A"),
					resource.TestCheckResourceAttr(resourceName, "value", "192.168.0.2"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						"dns_id":   knownvalue.NotNull(),
						"name":     knownvalue.StringExact("www2"),
						"type":     knownvalue.StringExact("A"),
						"value":    knownvalue.StringExact("192.168.0.2"),
						"ttl":      knownvalue.NotNull(),
						"priority": knownvalue.Null(),
						"weight":   knownvalue.Null(),
						"port":     knownvalue.Null(),
					}),
				},
			},
			{
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
			},
		},
	})
}

func testCheckSakuraDNSRecordDestroy(s *terraform.State) error {
	dnsOp := iaas.NewDNSOp(test.AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_dns_record" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		dnsID := rs.Primary.Attributes["dns_id"]
		if dnsID != "" {
			d, err := dnsOp.Read(context.Background(), common.SakuraCloudID(dnsID))
			if err != nil && !iaas.IsNotFoundError(err) {
				return fmt.Errorf("resource still exists: DNS: %s", rs.Primary.ID)
			}
			if d != nil {
				record := &iaas.DNSRecord{
					Name:  rs.Primary.Attributes["name"],
					Type:  types.EDNSRecordType(rs.Primary.Attributes["type"]),
					RData: rs.Primary.Attributes["value"],
					TTL:   utils.MustAtoI(rs.Primary.Attributes["ttl"]),
				}

				for _, r := range d.Records {
					if dns.IsSameDNSRecord(r, record) {
						return fmt.Errorf("resource still exists: DNSRecord: %s", rs.Primary.ID)
					}
				}
			}
		}
	}

	return nil
}

var testAccSakuraDNSRecord_basic = `
resource "sakura_dns" "foobar" {
  zone = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_dns_record" "foobar1" {
  dns_id = sakura_dns.foobar.id
  name   = "www"
  type   = "A"
  value  = "192.168.0.1"
}

resource "sakura_dns_record" "foobar2" {
  dns_id   = sakura_dns.foobar.id
  name     = "_sip._tls"
  type     = "SRV"
  value    = "www.sakura.ne.jp."
  priority = 1
  weight   = 2
  port     = 3
}
`

var testAccSakuraDNSRecord_update = `
resource "sakura_dns" "foobar" {
  zone = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_dns_record" "foobar1" {
  dns_id = sakura_dns.foobar.id
  name   = "www2"
  type   = "A"
  value  = "192.168.0.2"
}`

var testAccSakuraDNSRecord_withCount = `
resource "sakura_dns" "foobar" {
  zone = "{{ .arg0 }}"
}

variable "addresses" {
  default = ["192.168.0.1", "192.168.0.2"]
}

resource "sakura_dns_record" "foobar" {
  count  = 2
  dns_id = sakura_dns.foobar.id
  name   = "www"
  type   = "A"
  value  = var.addresses[count.index]
}`
