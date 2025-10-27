// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/eventbus-api-go"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceSchedule_basic(t *testing.T) {
	resourceName := "sakura_eventbus_schedule.foobar"
	rand := test.RandomName()
	var schedule v1.Schedule
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSchedule_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "recurring_step", "1"),
					resource.TestCheckResourceAttr(resourceName, "recurring_unit", "day"),
					resource.TestCheckResourceAttr(resourceName, "starts_at", "1700000000000"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraSchedule_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraScheduleExists(resourceName, &schedule),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag3"),
					resource.TestCheckResourceAttr(resourceName, "recurring_step", "20"),
					resource.TestCheckResourceAttr(resourceName, "recurring_unit", "min"),
					resource.TestCheckResourceAttr(resourceName, "starts_at", "1800000000000"),
				),
			},
		},
	})
}

func testCheckSakuraScheduleDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	scheduleOp := eventbus.NewScheduleOp(client.EventBusClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_eventbus_schedule" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := scheduleOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists Schedule: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraScheduleExists(n string, schedule *v1.Schedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Schedule ID is set")
		}

		client := test.AccClientGetter()
		scheduleOp := eventbus.NewScheduleOp(client.EventBusClient)

		foundSchedule, err := scheduleOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		foundID := strconv.FormatInt(foundSchedule.ID, 10)
		if foundID != rs.Primary.ID {
			return fmt.Errorf("not found Schedule: %s", rs.Primary.ID)
		}

		*schedule = *foundSchedule
		return nil
	}
}

var testAccSakuraSchedule_basic = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  simplenotification_access_token_wo        = "test"
  simplenotification_access_token_secret_wo = "test"
  credentials_wo_version                    = 1
}

resource "sakura_eventbus_schedule" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]

  process_configuration_id = sakura_eventbus_process_configuration.foobar.id
  recurring_step           = 1
  recurring_unit           = "day"
  starts_at                = 1700000000000
}`

var testAccSakuraSchedule_update = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  simplenotification_access_token_wo        = "test"
  simplenotification_access_token_secret_wo = "test"
  credentials_wo_version                    = 1
}

resource "sakura_eventbus_schedule" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  tags        = ["tag2", "tag3"]

  process_configuration_id = sakura_eventbus_process_configuration.foobar.id
  recurring_step           = 20
  recurring_unit           = "min"
  starts_at                = 1800000000000
}`
