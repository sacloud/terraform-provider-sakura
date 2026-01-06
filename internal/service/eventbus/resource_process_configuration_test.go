// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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
	var pc v1.CommonServiceItem
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
					resource.TestCheckNoResourceAttr(resourceName, "sakura_access_token_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "sakura_access_token_secret_wo"),
					resource.TestCheckResourceAttr(resourceName, "credentials_wo_version", "1"),
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
					resource.TestCheckNoResourceAttr(resourceName, "sakura_access_token_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "sakura_access_token_secret_wo"),
					resource.TestCheckResourceAttr(resourceName, "credentials_wo_version", "1"),
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
				Config:      test.BuildConfigWithArgs(testAccSakuraProcessConfiguration_validation_SimpleNotificationCredential, rand),
				ExpectError: regexp.MustCompile(`Expected "sakura_access_token_wo" to be configured`),
			},
			{
				Config:      test.BuildConfigWithArgs(testAccSakuraProcessConfiguration_validation_SimpleMQCredential, rand),
				ExpectError: regexp.MustCompile(`Expected "simplemq_api_key_wo" to be configured`),
			},
			{
				Config:             test.BuildConfigWithArgs(testAccSakuraProcessConfiguration_validation_SimpleMQCredentialWithVariables, rand),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
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

func testCheckSakuraProcessConfigurationExists(n string, pc *v1.CommonServiceItem) resource.TestCheckFunc {
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

		foundID := foundPC.ID
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

  sakura_access_token_wo        = "test"
  sakura_access_token_secret_wo = "test"
  credentials_wo_version        = 1
}`

var testAccSakuraProcessConfiguration_update = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"
  tags        = ["tag2", "tag3"]

  destination = "simplemq"
  parameters  = "{\"queue_name\": \"test-queue\", \"content\":\"TestContent\"}"

  simplemq_api_key_wo    = "test"
  credentials_wo_version = 1
}`

var testAccSakuraProcessConfiguration_validation_unknownDestination = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description-updated"

  destination = "unknown"
  parameters  = "{\"param\": \"something\"}"
}`

//nolint:gosec // hardcoded credentials but this is dummy data for test
var testAccSakuraProcessConfiguration_validation_SimpleNotificationCredential = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"

  destination = "simplenotification"
  parameters  = "{\"group_id\": \"123456789012\", \"message\":\"test message\"}"

  # missing -> sakura_access_token_wo
  sakura_access_token_secret_wo = "test"
  credentials_wo_version        = 1
}`

//nolint:gosec // hardcoded credentials but this is dummy data for test
var testAccSakuraProcessConfiguration_validation_SimpleMQCredential = `
resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"

  destination = "simplemq"
  parameters  = "{\"queue_name\": \"test-queue\", \"content\":\"TestContent\"}"

  # wrong credentials
  sakura_access_token_wo        = "test"
  sakura_access_token_secret_wo = "test"
  credentials_wo_version        = 1
}`

//nolint:gosec // hardcoded credentials but this is dummy data for test
var testAccSakuraProcessConfiguration_validation_SimpleMQCredentialWithVariables = `
// Using variable becomes unknown state in ValidateConfig
variable "simplemq_api_key_testvalue" {
  type = string
  default = "foo"
}

resource "sakura_eventbus_process_configuration" "foobar" {
  name        = "{{ .arg0 }}"

  destination            = "simplemq"
  parameters             = "{\"queue_name\": \"test-queue\", \"content\":\"TestContent\"}"
  simplemq_api_key_wo    = var.simplemq_api_key_testvalue
  credentials_wo_version = 1
}`
