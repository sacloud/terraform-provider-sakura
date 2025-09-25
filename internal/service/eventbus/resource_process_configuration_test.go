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

package eventbus_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/eventbus-api-go"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceProcessConfiguration_basic(t *testing.T) {
	resourceName := "sakura_eventbus_process_configuration.foobar"
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
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "destination", "simplenotification"),
					resource.TestCheckResourceAttr(resourceName, "parameters", "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"),
					resource.TestCheckNoResourceAttr(resourceName, "simplemq_api_key_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "simplenotification_access_token_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "simplenotification_access_token_secret_wo"),
					resource.TestCheckResourceAttr(resourceName, "simplenotification_credentials_wo_version", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "simplemq_credentials_wo_version"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraProcessConfiguration_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraProcessConfigurationExists(resourceName, &pc),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag3"),
					resource.TestCheckResourceAttr(resourceName, "destination", "simplemq"),
					resource.TestCheckResourceAttr(resourceName, "parameters", "{\"queue_name\": \"test-queue\", \"content\":\"TestContent\"}"),
					resource.TestCheckNoResourceAttr(resourceName, "simplemq_api_key_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "simplenotification_access_token_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "simplenotification_access_token_secret_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "simplenotification_credentials_wo_version"),
					resource.TestCheckResourceAttr(resourceName, "simplemq_credentials_wo_version", "1"),
				),
			},
		},
	})
}

func TestAccSakuraResourceProcessConfiguration_validation_credentials(t *testing.T) {
	rand := test.RandomName()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraProcessConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraProcessConfiguration_validation_unknownDestination, rand),
				ExpectError: regexp.MustCompile(`Unknown destination`),
			},
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraProcessConfiguration_validation_multipleVersion, rand),
				ExpectError: regexp.MustCompile(`"simplenotification_credentials_wo_version" is not necessary`),
			},
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraProcessConfiguration_validation_SimpleNotificationCredential, rand),
				ExpectError: regexp.MustCompile(`Expected "simplenotification_access_token_wo" to be configured`),
			},
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraProcessConfiguration_validation_SimpleMQCredential, rand),
				ExpectError: regexp.MustCompile(`Expected "simplemq_api_key_wo" to be configured`),
			},
		},
	})
}

func testCheckSakuraProcessConfigurationDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	processConfigurationOp := eventbus.NewProcessConfigurationOp(client.EventBusClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_eventbus_process_configuration" {
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
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1"]

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  simplenotification_access_token_wo        = "test"
  simplenotification_access_token_secret_wo = "test"
  simplenotification_credentials_wo_version = 1
}`

var testAccSakuraProcessConfiguration_update = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  tags        = ["tag2", "tag3"]

  destination = "simplemq"
  parameters  = "{\"queue_name\": \"test-queue\", \"content\":\"TestContent\"}"

  simplemq_api_key_wo             = "test"
  simplemq_credentials_wo_version = 1
}`

var testAccSakuraProcessConfiguration_validation_unknownDestination = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"

  destination = "unknown"
  parameters  = "{\"param\": \"something\"}"
}`

var testAccSakuraProcessConfiguration_validation_multipleVersion = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"

  destination = "simplemq"
  parameters  = "{\"queue_name\": \"test-queue\", \"content\":\"TestContent\"}"

  simplemq_api_key_wo                       = "test"
  simplemq_credentials_wo_version           = 1
  # unnecessary
  simplenotification_credentials_wo_version = 1
}`

var testAccSakuraProcessConfiguration_validation_SimpleNotificationCredential = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  # missing -> simplenotification_access_token_wo
  simplenotification_access_token_secret_wo       = "test"
  simplenotification_credentials_wo_version       = 1
}`

var testAccSakuraProcessConfiguration_validation_SimpleMQCredential = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"

  destination = "simplemq"
  parameters  = "{\"queue_name\": \"test-queue\", \"content\":\"TestContent\"}"

  # wrong credentials
  simplenotification_access_token_wo        = "test"
  simplenotification_access_token_secret_wo = "test"
  simplenotification_credentials_wo_version = 1
}`
