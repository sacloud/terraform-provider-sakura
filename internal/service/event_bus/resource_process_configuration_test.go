// Copyright 2016-2025 terraform-provider-sakura authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event_bus_test

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

func TestAccSakuraResourceProcessConfiguration_basic(t *testing.T) {
	resourceName := "sakura_event_bus_process_configuration.foobar"
	rand := test.RandomName()
	var pc v1.ProcessConfiguration
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraProcessConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraProcessConfiguration_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraProcessConfigurationExists(resourceName, &pc),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "destination", "simplenotification"),
					resource.TestCheckResourceAttr(resourceName, "parameters", "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"),
					resource.TestCheckResourceAttr(resourceName, "simplenotification_access_token", "test"),
					resource.TestCheckResourceAttr(resourceName, "simplenotification_access_token_secret", "test"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraProcessConfiguration_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraProcessConfigurationExists(resourceName, &pc),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "destination", "simplemq"),
					resource.TestCheckResourceAttr(resourceName, "parameters", "{\"queue_name\": \"test-queue\", \"content\":\"TestContent\"}"),
					resource.TestCheckResourceAttr(resourceName, "simplemq_api_key", "test"),
				),
			},
		},
	})
}

func testCheckSakuraProcessConfigurationDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	processConfigurationOp := eventbus.NewProcessConfigurationOp(client.EventBusClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_event_bus_process_configuration" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := processConfigurationOp.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists ProcessConfiguration: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraProcessConfigurationExists(n string, pc *v1.ProcessConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no ProcessConfiguration ID is set")
		}

		client := test.AccClientGetter()
		processConfigurationOp := eventbus.NewProcessConfigurationOp(client.EventBusClient)

		foundPC, err := processConfigurationOp.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		foundID := strconv.FormatInt(foundPC.ID, 10)
		if foundID != rs.Primary.ID {
			return fmt.Errorf("not found ProcessConfiguration: %s", rs.Primary.ID)
		}

		*pc = *foundPC
		return nil
	}
}

var testAccSakuraProcessConfiguration_basic = `
resource "sakura_event_bus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  simplenotification_access_token        = "test"
  simplenotification_access_token_secret = "test"
}`

var testAccSakuraProcessConfiguration_update = `
resource "sakura_event_bus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"

  destination = "simplemq"
  parameters  = "{\"queue_name\": \"test-queue\", \"content\":\"TestContent\"}"

	simplemq_api_key = "test"
}`
