// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraResourceAPIGWService_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION", "SAKURA_APIGW_SERVICE_HOST")

	resourceName := "sakura_apigw_service.foobar"
	rand := test.RandomName()
	host := os.Getenv("SAKURA_APIGW_SERVICE_HOST")
	var service v1.ServiceDetailResponse
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraAPIGWServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWService_basic, rand, host),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttr(resourceName, "port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "retries", "3"),
					resource.TestCheckResourceAttr(resourceName, "read_timeout", "60000"),
					resource.TestCheckResourceAttr(resourceName, "read_timeout", "60000"),
					resource.TestCheckResourceAttr(resourceName, "write_timeout", "60000"),
					resource.TestCheckResourceAttr(resourceName, "connect_timeout", "60000"),
					resource.TestCheckResourceAttr(resourceName, "authentication", "none"),
					resource.TestCheckResourceAttrSet(resourceName, "route_host"),
					resource.TestCheckNoResourceAttr(resourceName, "oidc"),
					resource.TestCheckNoResourceAttr(resourceName, "cors_config"),
					resource.TestCheckNoResourceAttr(resourceName, "object_storage_config"),
					resource.TestCheckResourceAttrPair(resourceName, "subscription_id", "sakura_apigw_subscription.foobar", "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWService_update, rand, host),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "https"),
					resource.TestCheckResourceAttr(resourceName, "path", "/api"),
					resource.TestCheckResourceAttr(resourceName, "port", "9080"),
					resource.TestCheckResourceAttr(resourceName, "retries", "7"),
					resource.TestCheckResourceAttr(resourceName, "read_timeout", "30000"),
					resource.TestCheckResourceAttr(resourceName, "write_timeout", "30000"),
					resource.TestCheckResourceAttr(resourceName, "connect_timeout", "30000"),
					resource.TestCheckResourceAttr(resourceName, "authentication", "none"),
					resource.TestCheckResourceAttrSet(resourceName, "route_host"),
					resource.TestCheckNoResourceAttr(resourceName, "oidc"),
					resource.TestCheckNoResourceAttr(resourceName, "cors_config"),
					resource.TestCheckNoResourceAttr(resourceName, "object_storage_config"),
					resource.TestCheckResourceAttrPair(resourceName, "subscription_id", "sakura_apigw_subscription.foobar", "id"),
				),
			},
		},
	})
}

func TestAccSakuraResourceAPIGWService_withConfigs(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION", "SAKURA_APIGW_SERVICE_HOST")

	resourceName := "sakura_apigw_service.foobar"
	rand := test.RandomName()
	host := os.Getenv("SAKURA_APIGW_SERVICE_HOST")
	var service v1.ServiceDetailResponse
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraAPIGWServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWService_withConfigs, rand, host),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "https"),
					resource.TestCheckResourceAttr(resourceName, "path", "/configs"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.access_control_allow_methods.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.access_control_allow_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.access_control_allow_methods.*", "POST"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.access_control_allow_headers", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.max_age", "3600"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.preflight_continue", "false"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.private_network", "false"),
					resource.TestCheckResourceAttr(resourceName, "object_storage_config.bucket", "test1"),
					resource.TestCheckResourceAttrSet(resourceName, "object_storage_config.region"),
					resource.TestCheckResourceAttrSet(resourceName, "object_storage_config.endpoint"),
					resource.TestCheckNoResourceAttr(resourceName, "object_storage_config.access_key_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "object_storage_config.secret_access_key_wo"),
					resource.TestCheckResourceAttr(resourceName, "object_storage_config.credentials_wo_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "object_storage_config.use_document_index", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "subscription_id", "sakura_apigw_subscription.foobar", "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWService_withConfigsUpdate, rand, host),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWServiceExists(resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "https"),
					resource.TestCheckResourceAttr(resourceName, "path", "/configs"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.access_control_allow_methods.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.access_control_allow_methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.access_control_allow_methods.*", "POST"),
					resource.TestCheckTypeSetElemAttr(resourceName, "cors_config.access_control_allow_methods.*", "OPTIONS"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.access_control_allow_headers", "X-*"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.access_control_allow_origins", "*"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.max_age", "1800"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.preflight_continue", "true"),
					resource.TestCheckResourceAttr(resourceName, "cors_config.private_network", "false"),
					resource.TestCheckResourceAttr(resourceName, "object_storage_config.bucket", "test2"),
					resource.TestCheckResourceAttrSet(resourceName, "object_storage_config.region"),
					resource.TestCheckResourceAttrSet(resourceName, "object_storage_config.endpoint"),
					resource.TestCheckNoResourceAttr(resourceName, "object_storage_config.access_key_wo"),
					resource.TestCheckNoResourceAttr(resourceName, "object_storage_config.secret_access_key_wo"),
					resource.TestCheckResourceAttr(resourceName, "object_storage_config.credentials_wo_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "object_storage_config.use_document_index", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "subscription_id", "sakura_apigw_subscription.foobar", "id"),
				),
			},
		},
	})
}

func testCheckSakuraAPIGWServiceDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	serviceOp := apigw.NewServiceOp(client.ApigwClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apigw_service" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := serviceOp.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists APIGW Service: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraAPIGWServiceExists(n string, service *v1.ServiceDetailResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no APIGW Service ID is set")
		}

		serviceOp := apigw.NewServiceOp(test.AccClientGetter().ApigwClient)
		foundService, err := serviceOp.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err != nil {
			return err
		}

		if foundService.ID.Value.String() != rs.Primary.ID {
			return fmt.Errorf("not found APIGW Service: %s", rs.Primary.ID)
		}

		*service = *foundService
		return nil
	}
}

var testSetupAPIGWService = testSetupAPIGWSub + `
resource "sakura_apigw_service" "foobar" {
  name     = "{{ .arg0 }}"
  tags     =  ["tag1"]
  protocol = "https"
  host     = "{{ .arg1 }}"
  subscription_id = sakura_apigw_subscription.foobar.id
}
`

var testAccSakuraAPIGWService_basic = testSetupAPIGWSub + `
resource "sakura_apigw_service" "foobar" {
  name     = "{{ .arg0 }}"
  tags     =  ["tag1"]
  protocol = "http"
  host     = "{{ .arg1 }}"
  port     = 8080
  retries  = 3
  subscription_id = sakura_apigw_subscription.foobar.id
}`

var testAccSakuraAPIGWService_update = testSetupAPIGWSub + `
resource "sakura_apigw_service" "foobar" {
  name     = "{{ .arg0 }}-updated"
  tags     = ["tag1", "tag2"]
  protocol = "https"
  host     = "{{ .arg1 }}"
  path     = "/api"
  port     = 9080
  retries  = 7
  read_timeout    = 30000
  write_timeout   = 30000
  connect_timeout = 30000
  subscription_id = sakura_apigw_subscription.foobar.id
}`

var testAccSakuraAPIGWService_withConfigs = testSetupAPIGWSub + `
data "sakura_object_storage_site" "foobar" {
  id = "isk01"
}

resource "sakura_apigw_service" "foobar" {
  name     = "{{ .arg0 }}"
  tags     =  ["tag1"]
  protocol = "https"
  host     = "{{ .arg1 }}"
  path     = "/configs"
  subscription_id = sakura_apigw_subscription.foobar.id
  cors_config = {
    access_control_allow_methods = ["GET", "POST"]
    access_control_allow_headers = "*"
    max_age       = 3600
  }
  object_storage_config = {
    bucket = "test1"
    region = data.sakura_object_storage_site.foobar.region
    endpoint = data.sakura_object_storage_site.foobar.s3_endpoint
	access_key_wo = "aaaaaaaaaaaaaaaaaaaa"
	secret_access_key_wo = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
    credentials_wo_version = 1
  }
}`

var testAccSakuraAPIGWService_withConfigsUpdate = testSetupAPIGWSub + `
data "sakura_object_storage_site" "foobar" {
  id = "isk01"
}

resource "sakura_apigw_service" "foobar" {
  name     = "{{ .arg0 }}-updated"
  tags     = ["tag1", "tag2"]
  protocol = "https"
  host     = "{{ .arg1 }}"
  path     = "/configs"
  subscription_id = sakura_apigw_subscription.foobar.id
  cors_config = {
    access_control_allow_methods = ["GET", "POST", "OPTIONS"]
    access_control_allow_headers = "X-*"
	access_control_allow_origins = "*"
    max_age = 1800
	preflight_continue = true
  }
  object_storage_config = {
    bucket = "test2"
    region = data.sakura_object_storage_site.foobar.region
    endpoint = data.sakura_object_storage_site.foobar.s3_endpoint
	access_key_wo = "aaaaaaaaaaaaaaaaaaaz"
	secret_access_key_wo = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbx"
    credentials_wo_version = 2
	use_document_index = false
  }
}`
