// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package enhanced_lb_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

const (
	envEnhancedLBRealServerIP0 = "SAKURACLOUD_ENHANCED_LB_SERVER0"
	envEnhancedLBRealServerIP1 = "SAKURACLOUD_ENHANCED_LB_SERVER1"
)

func TestAccSakuraEnhancedLB_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, envEnhancedLBRealServerIP0)

	resourceName := "sakura_enhanced_lb.foobar"
	rand := test.RandomName()
	ip := os.Getenv(envEnhancedLBRealServerIP0)
	subDomain := "acme-acctest1" + test.RandStringFromCharSet(5, "")
	elbDomain := os.Getenv(envEnhancedLBACMEDomain)

	var elb, elbUpd iaas.ProxyLB
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			test.CheckSakuraIconDestroy,
			testCheckSakuraEnhancedLBDestroy,
			test.CheckSakuraServerDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraEnhancedLB_basic, rand, ip),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraEnhancedLBExists(resourceName, &elb),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "plan", "100"),
					resource.TestMatchResourceAttr(resourceName, "fqdn", regexp.MustCompile(`.+\.sakura\.ne\.jp$`)),
					resource.TestCheckResourceAttr(resourceName, "vip_failover", "true"),
					resource.TestCheckResourceAttr(resourceName, "sticky_session", "true"),
					resource.TestCheckResourceAttr(resourceName, "gzip", "true"),
					resource.TestCheckResourceAttr(resourceName, "backend_http_keep_alive", "aggressive"),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol", "true"),
					resource.TestCheckResourceAttr(resourceName, "health_check.protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "health_check.delay_loop", "10"),
					resource.TestCheckResourceAttr(resourceName, "health_check.host_header", "usacloud.jp"),
					resource.TestCheckResourceAttr(resourceName, "health_check.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "sorry_server.ip_address", ip),
					resource.TestCheckResourceAttr(resourceName, "sorry_server.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "syslog.server", "133.242.0.1"),
					resource.TestCheckResourceAttr(resourceName, "syslog.port", "514"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.proxy_mode", "http"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.ssl_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.response_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.response_header.0.header", "Cache-Control"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.response_header.0.value", "public, max-age=10"),
					resource.TestCheckResourceAttr(resourceName, "server.0.port", "80"),
					resource.TestCheckResourceAttr(resourceName, "server.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "server.0.group", "group1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.host", "usacloud.jp"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.path", "/path"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source_ips", "192.0.2.1,192.0.2.2"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.group", "group1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.action", "fixed"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.fixed_status_code", "200"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.fixed_content_type", "text/plain"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.fixed_message_body", "example"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.redirect_status_code", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.0.redirect_location", ""),

					resource.TestCheckResourceAttr(resourceName, "rule.0.request_header_name", "foo"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.request_header_value", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.request_header_value_ignore_case", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.request_header_value_not_match", "true"),

					resource.TestCheckResourceAttr(resourceName, "certificate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate.0.common_name", subDomain+"."+elbDomain),
					resource.TestCheckResourceAttr(resourceName,
						"certificate.0.subject_alt_names",
						fmt.Sprintf("%s.%s, acme-acctest2.%s, acme-acctest3.%s", subDomain, elbDomain, elbDomain, elbDomain),
					),

					resource.TestCheckResourceAttrSet(resourceName, "vip"),
					resource.TestCheckResourceAttrPair(
						resourceName, "server.0.ip_address",
						"sakura_server.foobar", "ip_address",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "icon_id",
						"sakura_icon.foobar", "id",
					),
					resource.TestCheckResourceAttr(resourceName, "monitoring_suite.enabled", "true"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraEnhancedLB_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraEnhancedLBExists(resourceName, &elbUpd),
					func(state *terraform.State) error {
						if elb.ID == elbUpd.ID {
							return fmt.Errorf("sakura_enhanced_lb: plan wasn't updated")
						}
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "name", rand+"-upd"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-upd"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1-upd"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2-upd"),
					resource.TestCheckResourceAttr(resourceName, "plan", "500"),
					resource.TestCheckResourceAttr(resourceName, "sticky_session", "false"),
					resource.TestCheckResourceAttr(resourceName, "gzip", "false"),
					resource.TestCheckResourceAttr(resourceName, "backend_http_keep_alive", "safe"),
					resource.TestCheckResourceAttr(resourceName, "proxy_protocol", "false"),
					resource.TestCheckResourceAttr(resourceName, "health_check.protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "health_check.delay_loop", "20"),
					resource.TestCheckResourceAttr(resourceName, "health_check.host_header", ""),
					resource.TestCheckResourceAttr(resourceName, "health_check.path", ""),
					resource.TestCheckNoResourceAttr(resourceName, "sorry_server"),
					resource.TestCheckNoResourceAttr(resourceName, "syslog"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.proxy_mode", "https"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.ssl_policy", "TLS-1-3-2021-06"),
					resource.TestCheckResourceAttr(resourceName, "bind_port.0.response_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "server.0.port", "443"),
					resource.TestCheckResourceAttr(resourceName, "server.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "server.0.group", "group2"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.host", "upd.usacloud.jp"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.path", "/path-upd"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.source_ips", "192.0.2.11,192.0.2.12"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.group", "group2"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.action", "redirect"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.fixed_status_code", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.0.fixed_content_type", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.0.fixed_message_body", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.0.redirect_status_code", "301"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.redirect_location", "https://redirect.usacloud.jp"),
					resource.TestCheckResourceAttr(resourceName, "certificate.common_name", subDomain+"."+elbDomain),
					resource.TestCheckResourceAttr(resourceName,
						"certificate.0.subject_alt_names",
						fmt.Sprintf("%s.%s, acme-acctest2.%s, acme-acctest3.%s", subDomain, elbDomain, elbDomain, elbDomain),
					),
					resource.TestCheckResourceAttrSet(resourceName, "vip"),
					resource.TestCheckResourceAttrPair(
						resourceName, "server.0.ip_address",
						"sakura_server.foobar", "ip_address",
					),
					resource.TestCheckResourceAttr(resourceName, "monitoring_suite.enabled", "false"),
				),
			},
		},
	})
}

func testCheckSakuraEnhancedLBExists(n string, elb *iaas.ProxyLB) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no Enhanced LB ID is set")
		}

		elbOp := iaas.NewProxyLBOp(test.AccClientGetter())
		foundELB, err := elbOp.Read(context.Background(), common.SakuraCloudID(rs.Primary.ID))
		if err != nil {
			return err
		}

		if foundELB.ID.String() != rs.Primary.ID {
			return fmt.Errorf("not found Enhanced LB: %s", rs.Primary.ID)
		}
		*elb = *foundELB
		return nil
	}
}

func testCheckSakuraEnhancedLBDestroy(s *terraform.State) error {
	elbOp := iaas.NewProxyLBOp(test.AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_enhanced_lb" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := elbOp.Read(context.Background(), common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists Enhanced LB: %s", rs.Primary.ID)
		}
	}

	return nil
}

func TestAccImportSakuraEnhancedLB_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, envEnhancedLBRealServerIP0, envEnhancedLBRealServerIP1)

	ip0 := os.Getenv(envEnhancedLBRealServerIP0)
	ip1 := os.Getenv(envEnhancedLBRealServerIP1)
	rand := test.RandomName()

	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":                     rand,
			"vip_failover":             "true",
			"sticky_session":           "true",
			"timeout":                  "10",
			"region":                   "is1",
			"health_check.protocol":    "tcp",
			"health_check.delay_loop":  "20",
			"description":              "description",
			"tags.0":                   "tag1",
			"tags.1":                   "tag2",
			"bind_port.0.proxy_mode":   "https",
			"bind_port.0.port":         "443",
			"server.#":                 "2",
			"server.0.ip_address":      ip0,
			"server.0.port":            "80",
			"server.0.enabled":         "true",
			"server.1.ip_address":      ip1,
			"server.1.port":            "80",
			"server.1.enabled":         "true",
			"server.1.group":           "group1",
			"rule.0.action":            "forward",
			"rule.0.host":              "www.usacloud.jp",
			"rule.0.source_ips":        "192.0.2.1,192.0.2.2",
			"rule.0.path":              "/",
			"rule.0.group":             "group1",
			"monitoring_suite.enabled": "true",
		}

		if err := test.CompareStateMulti(s[0], expects); err != nil {
			return err
		}
		return test.StateNotEmptyMulti(s[0], "fqdn", "proxy_networks.0")
	}

	resourceName := "sakura_enhanced_lb.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraEnhancedLBDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraEnhancedLB_import, rand, ip0, ip1),
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

var testAccSakuraEnhancedLB_basic = `
resource "sakura_enhanced_lb" "foobar" {
  name           = "{{ .arg0 }}"
  plan           = 100
  vip_failover   = true
  sticky_session = true
  gzip           = true
  proxy_protocol = true
  timeout        = 10
  region         = "is1"

  backend_http_keep_alive = "aggressive"

  health_check = {
    protocol    = "http"
    delay_loop  = 10
    host_header = "usacloud.jp"
    path        = "/"
  }

  sorry_server = {
    ip_address = "{{ .arg1 }}"
    port       = 80
  }

  syslog = {
    server = "133.242.0.1"
    port   = 514
  }

  bind_port = [{
    proxy_mode = "http"
    port       = 80
    response_header = [{
      header = "Cache-Control"
      value  = "public, max-age=10"
    }]
  }]

  server = [{
    ip_address = sakura_server.foobar.ip_address
    port       = 80
    group      = "group1"
  }]

  rule = [{
	host               = "usacloud.jp"
	path               = "/path"
	source_ips         = "192.0.2.1,192.0.2.2"
	group              = "group1"
    action             = "fixed"
    fixed_status_code  = "200"
    fixed_content_type = "text/plain"
    fixed_message_body = "example"
    request_header_name = "foo"
    request_header_value = "1"
    request_header_value_ignore_case = true
    request_header_value_not_match = true
  }]

  monitoring_suite = {
    enabled = true
  }

  description = "description"
  tags        = ["tag1", "tag2"]
  icon_id     = sakura_icon.foobar.id
}

resource "sakura_icon" "foobar" {
  name          = "{{ .arg0 }}"
  base64content = "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"
}

resource "sakura_server" "foobar" {
  name = "{{ .arg0 }}"
  network_interface = [{
    upstream = "shared"
  }]
  force_shutdown = true
}

resource "sakura_enhanced_lb_acme" "foobar" {
  enhanced_lb_id               = sakura_enhanced_lb.foobar.id
  accept_tos                   = true
  common_name                  = "{{ .arg2 }}.{{ .arg1 }}"
  subject_alt_names            = ["acme-acctest2.{{ .arg1 }}", "acme-acctest3.{{ .arg1 }}"]
  update_delay_sec             = 120
  get_certificates_timeout_sec = 300
}
`

var testAccSakuraEnhancedLB_update = `
resource "sakura_enhanced_lb" "foobar" {
  name           = "{{ .arg0 }}-upd"
  plan           = 500
  vip_failover   = true
  sticky_session = false
  timeout        = 10
  region         = "is1"

  health_check = {
    protocol   = "tcp"
    delay_loop = 20
  }

  bind_port = {
    proxy_mode = "https"
    port       = 443
    ssl_policy = "TLS-1-3-2021-06"
  }

  server = [{
    ip_address = sakura_server.foobar.ip_address
    port       = 443
    group      = "group2"
  }]

  rule = [{
    host                 = "upd.usacloud.jp"
    path                 = "/path-upd"
    source_ips           = "192.0.2.11,192.0.2.12"
    group                = "group2"
    action               = "redirect"
    redirect_status_code = "301"
    redirect_location    = "https://redirect.usacloud.jp"
  }]

  monitoring_suite = {
    enabled = false
  }

  description = "description-upd"
  tags        = ["tag1-upd", "tag2-upd"]
}

resource "sakura_server" "foobar" {
  name = "{{ .arg0 }}"
  network_interface = [{
    upstream = "shared"
  }]
  force_shutdown = true
}
`

var testAccSakuraEnhancedLB_import = `
resource "sakura_enhanced_lb" "foobar" {
  name           = "{{ .arg0 }}"
  vip_failover   = true
  sticky_session = true
  timeout        = 10
  region         = "is1"
  health_check = {
    protocol   = "tcp"
    delay_loop = 20
  }
  bind_port = {
    proxy_mode = "https"
    port       = 443
  }
  server = [{
    ip_address = "{{ .arg1 }}"
    port       = 80
  },
  {
    ip_address = "{{ .arg2 }}"
    port       = 80
    group      = "group1"
  }]
  rule = [{
    host       = "www.usacloud.jp"
    path       = "/"
    source_ips = "192.0.2.1,192.0.2.2"
    group      = "group1"
  }]

  monitoring_suite = {
    enabled = true
  }

  description = "description"
  tags        = ["tag1", "tag2"]
}
`
