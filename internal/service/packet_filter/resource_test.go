// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package packet_filter_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraPacketFilter_basic(t *testing.T) {
	resourceName := "sakura_packet_filter.foobar"
	rand := test.RandomName()

	var filter iaas.PacketFilter
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraPacketFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraPacketFilter_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraPacketFilterExists(resourceName, &filter),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraPacketFilter_update, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
				),
			},
		},
	})
}

func testCheckSakuraPacketFilterExists(n string, filter *iaas.PacketFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no PacketFilter ID is set")
		}

		client := test.AccClientGetter()
		zone := rs.Primary.Attributes["zone"]
		pfOp := iaas.NewPacketFilterOp(client)

		foundPacketFilter, err := pfOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))
		if err != nil {
			return err
		}

		if foundPacketFilter.ID.String() != rs.Primary.ID {
			return fmt.Errorf("not found PacketFilter: %s", rs.Primary.ID)
		}

		*filter = *foundPacketFilter
		return nil
	}
}

func testCheckSakuraPacketFilterDestroy(s *terraform.State) error {
	pfOp := iaas.NewPacketFilterOp(test.AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_packet_filter" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		zone := rs.Primary.Attributes["zone"]
		_, err := pfOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))

		if err == nil {
			return fmt.Errorf("still exists PacketFilter: %s", rs.Primary.ID)
		}
	}

	return nil
}

var testAccSakuraPacketFilter_basic = `
resource "sakura_packet_filter" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
}`

var testAccSakuraPacketFilter_update = `
resource "sakura_packet_filter" "foobar" {
  name        = "{{ .arg0 }}-upd"
  description = "description-upd"
}`
