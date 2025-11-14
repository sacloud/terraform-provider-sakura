// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/eventbus-api-go"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceTrigger_basic(t *testing.T) {
	resourceName := "sakura_eventbus_trigger.foobar"
	rand := test.RandomName()
	var trigger v1.CommonServiceItem
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraTrigger_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttrPair(resourceName, "process_configuration_id",
						"sakura_eventbus_process_configuration.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "source", "test-source"),
					resource.TestCheckResourceAttr(resourceName, "types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "types.0", "type1"),
					resource.TestCheckResourceAttr(resourceName, "types.1", "type2"),
					resource.TestCheckResourceAttr(resourceName, "conditions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.op", "eq"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.values.0", "value1"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraTrigger_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag3"),
					resource.TestCheckResourceAttrPair(resourceName, "process_configuration_id",
						"sakura_eventbus_process_configuration.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "source", "test-source-updated"),
					resource.TestCheckResourceAttr(resourceName, "types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "types.0", "type2"),
					resource.TestCheckResourceAttr(resourceName, "types.1", "type3"),
					resource.TestCheckResourceAttr(resourceName, "conditions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.key", "key2"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.op", "in"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.values.0", "value1"),
					resource.TestCheckResourceAttr(resourceName, "conditions.0.values.1", "value2"),
					resource.TestCheckResourceAttr(resourceName, "conditions.1.key", "key3"),
					resource.TestCheckResourceAttr(resourceName, "conditions.1.op", "eq"),
					resource.TestCheckResourceAttr(resourceName, "conditions.1.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "conditions.1.values.0", "value3"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraTrigger_onlyRequired, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraTriggerExists(resourceName, &trigger),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag3"),
					resource.TestCheckResourceAttrPair(resourceName, "process_configuration_id",
						"sakura_eventbus_process_configuration.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "source", "test-source-updated"),
					resource.TestCheckNoResourceAttr(resourceName, "types"),
					resource.TestCheckNoResourceAttr(resourceName, "conditions"),
				),
			},
		},
	})
}

func testCheckSakuraTriggerDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	triggerOp := eventbus.NewTriggerOp(client.EventBusClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_eventbus_trigger" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := triggerOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists Trigger: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraTriggerExists(n string, trigger *v1.CommonServiceItem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Trigger ID is set")
		}

		client := test.AccClientGetter()
		triggerOp := eventbus.NewTriggerOp(client.EventBusClient)

		foundTrigger, err := triggerOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		foundID := foundTrigger.ID
		if foundID != rs.Primary.ID {
			return fmt.Errorf("not found Trigger: %s", rs.Primary.ID)
		}

		*trigger = *foundTrigger
		return nil
	}
}

var testAccSakuraTrigger_basic = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  sakura_access_token_wo        = "test"
  sakura_access_token_secret_wo = "test"
  credentials_wo_version        = 1
}

resource "sakura_eventbus_trigger" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]

  process_configuration_id = sakura_eventbus_process_configuration.foobar.id
  source                   = "test-source"
  types                    = ["type1", "type2"]
  conditions               = [
    {
      key    = "key1"
      op     = "eq"
      values = ["value1"]
    },
  ]
}`

var testAccSakuraTrigger_update = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  sakura_access_token_wo        = "test"
  sakura_access_token_secret_wo = "test"
  credentials_wo_version        = 1
}

resource "sakura_eventbus_trigger" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  tags        = ["tag2", "tag3"]

  process_configuration_id = sakura_eventbus_process_configuration.foobar.id
  source                   = "test-source-updated"
  types                    = ["type2", "type3"]
  conditions               = [
    {
      key    = "key2"
      op     = "in"
      values = ["value1", "value2"]
    },
    {
      key    = "key3"
      op     = "eq"
      values = ["value3"]
    },
  ]
}`

var testAccSakuraTrigger_onlyRequired = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  sakura_access_token_wo        = "test"
  sakura_access_token_secret_wo = "test"
  credentials_wo_version        = 1
}

resource "sakura_eventbus_trigger" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  tags        = ["tag2", "tag3"]

  process_configuration_id = sakura_eventbus_process_configuration.foobar.id
  source                   = "test-source-updated"
}`
