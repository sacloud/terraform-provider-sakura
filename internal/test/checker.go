// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

func CheckSakuraDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource is not exists: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("id is not set: %s", n)
		}
		return nil
	}
}

func CheckSakuravSwitchDestroy(s *terraform.State) error {
	swOp := iaas.NewSwitchOp(AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_vswitch" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		zone := rs.Primary.Attributes["zone"]
		_, err := swOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("resource vSwitch[%s] still exists", rs.Primary.ID)
		}
	}

	return nil
}

func CheckSakuraIconDestroy(s *terraform.State) error {
	iconOp := iaas.NewIconOp(AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_icon" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := iconOp.Read(context.Background(), common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists Icon: %s", rs.Primary.ID)
		}
	}

	return nil
}

func CheckSakuraServerDestroy(s *terraform.State) error {
	serverOp := iaas.NewServerOp(AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_server" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		zone := rs.Primary.Attributes["zone"]
		_, err := serverOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))

		if err == nil {
			return fmt.Errorf("still exists Server:%s", rs.Primary.ID)
		}
	}

	return nil
}

func CheckSakuraCloudDiskDestroy(s *terraform.State) error {
	diskOp := iaas.NewDiskOp(AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_disk" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		zone := rs.Primary.Attributes["zone"]
		_, err := diskOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))

		if err == nil {
			return fmt.Errorf("still exists Disk[%s]", rs.Primary.ID)
		}
	}

	return nil
}

func CheckSakuraCloudDNSRecordDestroy(s *terraform.State) error {
	dnsOp := iaas.NewDNSOp(AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_dns_record" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		dnsID := rs.Primary.Attributes["dns_id"]
		if dnsID != "" {
			dns, err := dnsOp.Read(context.Background(), common.SakuraCloudID(dnsID))
			if err != nil && !iaas.IsNotFoundError(err) {
				return fmt.Errorf("resource still exists: DNS: %s", rs.Primary.ID)
			}
			if dns != nil {
				record := &iaas.DNSRecord{
					Name:  rs.Primary.Attributes["name"],
					Type:  types.EDNSRecordType(rs.Primary.Attributes["type"]),
					RData: rs.Primary.Attributes["value"],
					TTL:   common.MustAtoI(rs.Primary.Attributes["ttl"]),
				}

				for _, r := range dns.Records {
					if isSameDNSRecord(r, record) {
						return fmt.Errorf("resource still exists: DNSRecord: %s", rs.Primary.ID)
					}
				}
			}
		}
	}

	return nil
}

func isSameDNSRecord(r1, r2 *iaas.DNSRecord) bool {
	return r1.Name == r2.Name && r1.RData == r2.RData && r1.TTL == r2.TTL && r1.Type == r2.Type
}
