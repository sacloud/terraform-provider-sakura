// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package bridge_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraBridge_basic(t *testing.T) {
	resourceName := "sakura_bridge.foobar"
	rand := test.RandomName()

	var bridge iaas.Bridge
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraBridgeDestroy,
			test.CheckSakuravSwitchDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraBridge_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraBridgeExists(resourceName, &bridge),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraBridge_disconnectvSwitch, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraBridgeExists(resourceName, &bridge),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
				),
			},
		},
	})
}

func testCheckSakuraBridgeExists(n string, bridge *iaas.Bridge) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no bridge ID is set")
		}

		client := test.AccClientGetter()
		bridgeOp := iaas.NewBridgeOp(client)
		zone := rs.Primary.Attributes["zone"]
		foundBridge, err := bridgeOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))

		if err != nil {
			return err
		}

		if foundBridge.ID.String() != rs.Primary.ID {
			return errors.New("bridge not found")
		}

		*bridge = *foundBridge

		return nil
	}
}

func testCheckSakuraBridgeDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_bridge" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		bridgeOp := iaas.NewBridgeOp(client)
		zone := rs.Primary.Attributes["zone"]
		_, err := bridgeOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))

		if err == nil {
			return errors.New("bridge still exists")
		}
	}

	return nil
}

func TestAccImportSakuraBridge_basic(t *testing.T) {
	rand := test.RandomName()
	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":        rand,
			"description": "description",
			"zone":        os.Getenv("SAKURACLOUD_ZONE"),
		}

		return test.CompareStateMulti(s[0], expects)
	}

	resourceName := "sakura_bridge.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraBridgeDestroy,
			test.CheckSakuravSwitchDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraBridge_basic, rand),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateCheck:  checkFn,
				ImportStateVerify: true,
			},
		},
	})
}

var testAccSakuraBridge_basic = `
resource "sakura_vswitch" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  bridge_id   = sakura_bridge.foobar.id
}
resource "sakura_bridge" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
}`

var testAccSakuraBridge_disconnectvSwitch = `
resource "sakura_bridge" "foobar" {
  name        = "{{ .arg0 }}-upd"
  description = "description-upd"
}`
