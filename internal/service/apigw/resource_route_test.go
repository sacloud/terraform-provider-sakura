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

func TestAccSakuraResourceAPIGWRoute_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_APIGW_NO_SUBSCRIPTION", "SAKURA_APIGW_SERVICE_HOST")

	resourceName := "sakura_apigw_route.foobar"
	rand := test.RandomName()
	host := os.Getenv("SAKURA_APIGW_SERVICE_HOST")
	var route v1.RouteDetail
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraAPIGWRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWRoute_basic, rand, host),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "protocols", "https"),
					resource.TestCheckResourceAttr(resourceName, "path", "/"+rand),
					resource.TestCheckResourceAttr(resourceName, "hosts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "methods.#", "9"),
					resource.TestCheckTypeSetElemAttr(resourceName, "methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "methods.*", "POST"),
					resource.TestCheckTypeSetElemAttr(resourceName, "methods.*", "PUT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "methods.*", "DELETE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "methods.*", "PATCH"),
					resource.TestCheckTypeSetElemAttr(resourceName, "methods.*", "OPTIONS"),
					resource.TestCheckTypeSetElemAttr(resourceName, "methods.*", "HEAD"),
					resource.TestCheckTypeSetElemAttr(resourceName, "methods.*", "CONNECT"),
					resource.TestCheckTypeSetElemAttr(resourceName, "methods.*", "TRACE"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.protocols", "https"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.restricted_by", "denyIps"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.0", "192.168.0.10"),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.id", "sakura_apigw_group.foobar", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "groups.0.name"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.http_method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.allow.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.allow.body.0", "foo"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.header_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.header_keys.0", "X-Remove-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.query_params.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.query_params.0", "remove_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.body.0", "remove_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.headers.0.from", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.headers.0.to", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.query_params.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.query_params.0.from", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.query_params.0.to", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.body.0.from", "old_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.body.0.to", "new_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.headers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.headers.0.key", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.headers.0.value", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.headers.1.key", "X-Old2-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.headers.1.value", "X-New2-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.query_params.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.query_params.0.key", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.query_params.0.value", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.body.0.key", "old_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.body.0.value", "new_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.add.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.add.headers.0.key", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.add.headers.0.value", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.add.query_params.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.add.query_params.0.key", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.add.query_params.0.value", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.add.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.add.body.0.key", "old_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.add.body.0.value", "new_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.headers.0.key", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.headers.0.value", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.query_params.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.query_params.0.key", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.query_params.0.value", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.body.0.key", "old_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.body.0.value", "new_body"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.allow.json_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.allow.json_keys.0", "foo"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.remove.if_status_code.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.remove.if_status_code.0", "426"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.remove.header_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.remove.header_keys.0", "X-Remove-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.remove.json_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.remove.json_keys.0", "remove_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.rename.if_status_code.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.rename.if_status_code.0", "426"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.rename.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.rename.headers.0.from", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.rename.headers.0.to", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.replace.if_status_code.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.replace.if_status_code.0", "426"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.replace.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.replace.headers.0.key", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.replace.headers.0.value", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.replace.json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.replace.json.0.key", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.replace.json.0.value", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.replace.body", "new_body_content"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.if_status_code.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.if_status_code.0", "426"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.headers.0.key", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.headers.0.value", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.json.0.key", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.json.0.value", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.if_status_code.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.if_status_code.0", "426"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.headers.0.key", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.headers.0.value", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.json.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.json.0.key", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.json.0.value", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.json.1.key", "old_param2"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.json.1.value", "new_param2"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraAPIGWRoute_update, rand, host),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraAPIGWRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "protocols", "http,https"),
					resource.TestCheckResourceAttr(resourceName, "path", "/"+rand+"2"),
					resource.TestCheckResourceAttr(resourceName, "hosts.#", "1"), // maybe only 1 auto-issued host is returned from APIGW
					resource.TestCheckResourceAttr(resourceName, "methods.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "methods.0", "GET"),
					resource.TestCheckResourceAttr(resourceName, "methods.1", "HEAD"),
					resource.TestCheckResourceAttr(resourceName, "methods.2", "OPTIONS"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.protocols", "http,https"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.restricted_by", "allowIps"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_restriction.ips.0", "192.168.0.1"),
					resource.TestCheckNoResourceAttr(resourceName, "groups"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.http_method", "POST"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.allow.body.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "request_transformation.allow.body.*", "foo"),
					resource.TestCheckTypeSetElemAttr(resourceName, "request_transformation.allow.body.*", "bar"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.header_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.header_keys.0", "X-Remove-Header"),
					resource.TestCheckNoResourceAttr(resourceName, "request_transformation.remove.query_params"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.body.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.body.0", "remove_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.remove.body.1", "remove_body2"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.headers.0.from", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.headers.0.to", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.query_params.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.query_params.0.from", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.query_params.0.to", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.body.0.from", "old_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.rename.body.0.to", "new_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.headers.0.key", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.headers.0.value", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.query_params.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.query_params.0.key", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.query_params.0.value", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.body.0.key", "old_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.replace.body.0.value", "new_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.headers.0.key", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.headers.0.value", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.query_params.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.query_params.0.key", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.query_params.0.value", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.body.0.key", "old_body"),
					resource.TestCheckResourceAttr(resourceName, "request_transformation.append.body.0.value", "new_body"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.allow.json_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.allow.json_keys.0", "foo"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.remove.if_status_code.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.remove.if_status_code.0", "429"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.remove.header_keys.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.remove.header_keys.0", "X-Remove-Header"),
					resource.TestCheckNoResourceAttr(resourceName, "response_transformation.remove.json_keys"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.if_status_code.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.if_status_code.0", "426"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.headers.0.key", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.headers.0.value", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.json.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.json.0.key", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.json.0.value", "new_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.json.1.key", "old2_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.add.json.1.value", "new2_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.if_status_code.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.if_status_code.0", "426"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.headers.0.key", "X-Old-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.headers.0.value", "X-New-Header"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.json.0.key", "old_param"),
					resource.TestCheckResourceAttr(resourceName, "response_transformation.append.json.0.value", "new_param"),
				),
			},
		},
	})
}

func testCheckSakuraAPIGWRouteDestroy(s *terraform.State) error {
	client := test.AccClientGetter()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_apigw_route" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		routeOp := apigw.NewRouteOp(client.ApigwClient, uuid.MustParse(rs.Primary.Attributes["service_id"]))
		_, err := routeOp.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists APIGW Route: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraAPIGWRouteExists(n string, route *v1.RouteDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no APIGW Service ID is set")
		}

		client := test.AccClientGetter().ApigwClient
		routeOp := apigw.NewRouteOp(client, uuid.MustParse(rs.Primary.Attributes["service_id"]))
		foundRoute, err := routeOp.Read(context.Background(), uuid.MustParse(rs.Primary.ID))
		if err != nil {
			return err
		}

		if foundRoute.ID.Value.String() != rs.Primary.ID {
			return fmt.Errorf("not found APIGW Route: %s", rs.Primary.ID)
		}

		*route = *foundRoute
		return nil
	}
}

var testAccSakuraAPIGWRoute_basic = testSetupAPIGWService + `
resource "sakura_apigw_group" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1"]
}

resource "sakura_apigw_route" "foobar" {
  name       = "{{ .arg0 }}"
  tags       = ["tag1"]
  service_id = sakura_apigw_service.foobar.id
  protocols  = "https"
  path       = "/{{ .arg0 }}"
  ip_restriction = {
    protocols = "https"
    restricted_by = "denyIps"
    ips = ["192.168.0.10"]
  }
  groups = [{
    id = sakura_apigw_group.foobar.id,
    enabled = true
  }]
  request_transformation = {
    http_method = "GET",
    allow = {
      body = ["foo"],
    },
    remove = {
      header_keys = ["X-Remove-Header"],
      query_params = ["remove_param"],
      body = ["remove_body"]
    },
    rename = {
      headers = [{
        from = "X-Old-Header",
        to   = "X-New-Header"
      }]
      query_params = [{
        from = "old_param",
        to   = "new_param"
      }]
      body = [{
        from = "old_body",
        to   = "new_body"
      }]
    },
    replace = {
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      },{
        key   = "X-Old2-Header",
        value = "X-New2-Header"
      }]
      query_params = [{
        key   = "old_param",
        value = "new_param"
      }]
      body = [{
        key   = "old_body",
        value = "new_body"
      }]
    },
    add = {
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      }]
      query_params = [{
        key   = "old_param",
        value = "new_param"
      }]
      body = [{
        key   = "old_body",
        value = "new_body"
      }]
    },
    append = {
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      }]
      query_params = [{
        key   = "old_param",
        value = "new_param"
      }]
      body = [{
        key   = "old_body",
        value = "new_body"
      }]
    }
  }
  response_transformation = {
    allow = {
      json_keys = ["foo"],
    },
    remove = {
      if_status_code = [426]
      header_keys = ["X-Remove-Header"],
      json_keys = ["remove_param"],
    },
    rename = {
      if_status_code = [426]
      headers = [{
        from = "X-Old-Header",
        to   = "X-New-Header"
      }]
    },
    replace = {
      if_status_code = [426]
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      }]
      json = [{
        key   = "old_param",
        value = "new_param"
      }]
      body = "new_body_content"
    },
    add = {
      if_status_code = [426]
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      }]
      json = [{
        key   = "old_param",
        value = "new_param"
      }]
    },
    append = {
      if_status_code = [426]
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      }]
      json = [{
        key   = "old_param",
        value = "new_param"
      },{
        key   = "old_param2",
        value = "new_param2"
      }]
    }
  }
}`

var testAccSakuraAPIGWRoute_update = testSetupAPIGWService + `
resource "sakura_apigw_group" "foobar" {
  name = "{{ .arg0 }}"
  tags = ["tag1"]
}

resource "sakura_apigw_route" "foobar" {
  name       = "{{ .arg0 }}-updated"
  tags       = ["tag1", "tag2"]
  service_id = sakura_apigw_service.foobar.id
  protocols  = "http,https"
  path       = "/{{ .arg0 }}2"
  methods    = ["GET","HEAD","OPTIONS"]
  ip_restriction = {
    protocols = "http,https"
    restricted_by = "allowIps"
    ips = ["192.168.0.1"]
  }
  request_transformation = {
    http_method = "POST",
    allow = {
      body = ["foo", "bar"],
    },
    remove = {
      header_keys = ["X-Remove-Header"],
      body = ["remove_body", "remove_body2"]
    },
    rename = {
      headers = [{
        from = "X-Old-Header",
        to   = "X-New-Header"
      }]
      query_params = [{
        from = "old_param",
        to   = "new_param"
      }]
      body = [{
        from = "old_body",
        to   = "new_body"
      }]
    },
    replace = {
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      }]
      query_params = [{
        key   = "old_param",
        value = "new_param"
      }]
      body = [{
        key   = "old_body",
        value = "new_body"
      }]
    },
    append = {
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      }]
      query_params = [{
        key   = "old_param",
        value = "new_param"
      }]
      body = [{
        key   = "old_body",
        value = "new_body"
      }]
    }
  }
  response_transformation = {
    allow = {
      json_keys = ["foo"],
    },
    remove = {
      if_status_code = [429]
      header_keys = ["X-Remove-Header"],
    },
    add = {
      if_status_code = [426]
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      }]
      json = [{
        key   = "old_param",
        value = "new_param"
      },{
        key   = "old2_param",
        value = "new2_param"
      }]
    },
    append = {
      if_status_code = [426]
      headers = [{
        key   = "X-Old-Header",
        value = "X-New-Header"
      }]
      json = [{
        key   = "old_param",
        value = "new_param"
      }]
    }
  }
}`
