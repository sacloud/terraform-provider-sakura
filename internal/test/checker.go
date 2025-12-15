// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
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

func CheckSakuraDataSourceNotExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if ok && rs.Primary.ID != "" {
			return fmt.Errorf("resource still exists: %s", n)
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

func CheckSakuraInternetDestroy(s *terraform.State) error {
	internetOp := iaas.NewInternetOp(AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_internet" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		zone := rs.Primary.Attributes["zone"]
		_, err := internetOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("resource Internet(switch+router)[%s] still exists:", rs.Primary.ID)
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

func CheckSakuraServerExists(n string, server *iaas.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Server ID is set")
		}

		client := AccClientGetter()
		serverOp := iaas.NewServerOp(client)
		zone := rs.Primary.Attributes["zone"]

		foundServer, err := serverOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))
		if err != nil {
			return err
		}

		if foundServer.ID.String() != rs.Primary.ID {
			return fmt.Errorf("not found Server: %s", rs.Primary.ID)
		}

		*server = *foundServer
		return nil
	}
}

func CheckSakuraServerAttributes(server *iaas.Server) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !server.InstanceStatus.IsUp() {
			return fmt.Errorf("unexpected server status: status=%v", server.InstanceStatus)
		}

		if len(server.Interfaces) == 0 {
			return errors.New("unexpected server NIC status: interfaces is nil")
		}

		if server.Interfaces[0].SwitchID.IsEmpty() || server.Interfaces[0].SwitchScope != types.Scopes.Shared {
			return fmt.Errorf("unexpected server NIC status: %#v", server.Interfaces[0])
		}

		return nil
	}
}

func CheckSakuraDiskDestroy(s *terraform.State) error {
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

func CheckSakuraDNSDestroy(s *terraform.State) error {
	dnsOp := iaas.NewDNSOp(AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_dns" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := dnsOp.Read(context.Background(), common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("resource still exists: DNS: %s", rs.Primary.ID)
		}
	}

	return nil
}

func CheckSakuraDNSRecordDestroy(s *terraform.State) error {
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
