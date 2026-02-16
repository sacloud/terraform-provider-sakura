// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ipv4_ptr_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

const (
	envTestDomain = "SAKURA_IPV4_PTR_DOMAIN"
)

var (
	testDomain string
)

func TestAccSakuraIPv4Ptr_basic(t *testing.T) {
	test.SkipIfFakeModeEnabled(t)

	var ip iaas.IPAddress
	if domain, ok := os.LookupEnv(envTestDomain); ok {
		testDomain = domain
	} else {
		t.Skipf("ENV %q is required. skip", envTestDomain)
		return
	}
	rand := test.RandomName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraIPv4PtrDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraIPv4PtrConfig_basic, rand, testDomain),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIPv4PtrExists("sakura_ipv4_ptr.foobar", &ip),
					resource.TestCheckResourceAttr(
						"sakura_ipv4_ptr.foobar", "hostname", fmt.Sprintf("terraform-test-domain03.%s", testDomain)),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccCheckSakuraIPv4PtrConfig_update, rand, testDomain),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraIPv4PtrExists("sakura_ipv4_ptr.foobar", &ip),
					resource.TestCheckResourceAttr(
						"sakura_ipv4_ptr.foobar", "hostname", fmt.Sprintf("terraform-test-domain04.%s", testDomain)),
				),
			},
		},
	})
}

func testCheckSakuraIPv4PtrExists(n string, ip *iaas.IPAddress) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no IPv4Ptr ID is set")
		}

		client := test.AccClientGetter()
		zone := rs.Primary.Attributes["zone"]
		ipAddrOp := iaas.NewIPAddressOp(client)

		foundIPv4Ptr, err := ipAddrOp.Read(context.Background(), zone, rs.Primary.ID)
		if err != nil {
			return err
		}

		if foundIPv4Ptr.IPAddress != rs.Primary.ID {
			return fmt.Errorf("not found IPv4Ptr: %s", rs.Primary.ID)
		}
		if foundIPv4Ptr.HostName == "" {
			return fmt.Errorf("hostname is empty IPv4Ptr: %s", foundIPv4Ptr.IPAddress)
		}

		*ip = *foundIPv4Ptr
		return nil
	}
}

func testCheckSakuraIPv4PtrDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	ipAddrOp := iaas.NewIPAddressOp(client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_ipv4_ptr" {
			continue
		}

		zone := rs.Primary.Attributes["zone"]
		ip, err := ipAddrOp.Read(context.Background(), zone, rs.Primary.ID)

		if err == nil && ip.HostName != "" {
			return fmt.Errorf("still exists IPv4Ptr: %s", ip.IPAddress)
		}
	}

	return nil
}

var testAccCheckSakuraIPv4PtrConfig_basic = `
data sakura_dns "dns" {
  name = "{{ .arg1 }}"
}

resource sakura_dns_record "record01" {
  dns_id = data.sakura_dns.dns.id
  name   = "terraform-test-domain03"
  type   = "A"
  value  = sakura_server.server.ip_address
  ttl    = 10
}

data sakura_archive "ubuntu" {
  os_type = "ubuntu"
}

resource sakura_disk "foobar" {
  name              = "{{ .arg0 }}"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource sakura_server "server" {
  name  = "{{ .arg0 }}"
  disks = [sakura_disk.foobar.id]
  network_interface = [{
    upstream = "shared"
  }]
}

resource "sakura_ipv4_ptr" "foobar" {
  ip_address = sakura_server.server.ip_address
  hostname   = "terraform-test-domain03.{{ .arg1 }}"
  retry_max  = 100
}`

var testAccCheckSakuraIPv4PtrConfig_update = `
data sakura_dns "dns" {
  name = "{{ .arg1 }}"
}

resource sakura_dns_record "record01" {
  dns_id = data.sakura_dns.dns.id
  name   = "terraform-test-domain04"
  type   = "A"
  value  = sakura_server.server.ip_address
  ttl    = 10
}

data sakura_archive "ubuntu" {
  os_type = "ubuntu"
}

resource sakura_disk "foobar" {
  name              = "{{ .arg0 }}"
  source_archive_id = data.sakura_archive.ubuntu.id
}

resource sakura_server "server" {
  name  = "{{ .arg0 }}"
  disks = [sakura_disk.foobar.id]
  network_interface = [{
    upstream = "shared"
  }]
}

resource "sakura_ipv4_ptr" "foobar" {
  ip_address = sakura_server.server.ip_address
  hostname   = "terraform-test-domain04.{{ .arg1 }}"
  retry_max  = 100
}`
