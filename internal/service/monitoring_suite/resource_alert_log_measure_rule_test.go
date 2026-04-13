// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	v1 "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteAlertLogMeasureRule_basic(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID", "SAKURA_MONITORING_SUITE_LOG_STORAGE_ID")

	resourceName := "sakura_monitoring_suite_alert_log_measure_rule.foobar"
	name := strings.ToLower(test.RandomName())
	lsId := os.Getenv("SAKURA_MONITORING_SUITE_LOG_STORAGE_ID")
	msId := os.Getenv("SAKURA_MONITORING_SUITE_METRIC_STORAGE_ID")
	expectdMatchers1 := `[{"type":"string","operator":"eq","field":"service_name","value":"example","value_list":["example"]}]`
	expectdMatchers2 := `[{"type":"or","matchers":[{"type":"string","field":"text_payload","value":"value","operator":"eq","value_list":[]},{"type":"number","field":"http_request_size","value":1000,"operator":"gt","value_list":[]}]}]`

	var routing v1.LogMeasureRule
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteAlertLogMeasureRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertLogMeasureRule_basic, name, lsId, msId),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteAlertLogMeasureRuleExists(resourceName, &routing),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_project_id", "sakura_monitoring_suite_alert_project.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "log_storage_id", lsId),
					resource.TestCheckResourceAttr(resourceName, "metric_storage_id", msId),
					resource.TestCheckResourceAttr(resourceName, "rule.version", "v1"),
					resource.TestCheckResourceAttrWith(resourceName, "rule.query.matchers", func(value string) error {
						var matchers, expected []map[string]interface{}
						if err := json.Unmarshal([]byte(value), &matchers); err != nil {
							return err
						}
						if err := json.Unmarshal([]byte(expectdMatchers1), &expected); err != nil {
							return err
						}
						if cmp.Equal(matchers, expected) {
							return nil
						} else {
							return fmt.Errorf("got invalid matchers: %#v", matchers)
						}
					}),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteAlertLogMeasureRule_update, name, lsId, msId),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteAlertLogMeasureRuleExists(resourceName, &routing),
					resource.TestCheckResourceAttr(resourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttrPair(resourceName, "alert_project_id", "sakura_monitoring_suite_alert_project.foobar", "id"),
					resource.TestCheckResourceAttr(resourceName, "log_storage_id", lsId),
					resource.TestCheckResourceAttr(resourceName, "metric_storage_id", msId),
					resource.TestCheckResourceAttr(resourceName, "rule.version", "v1"),
					resource.TestCheckResourceAttrWith(resourceName, "rule.query.matchers", func(value string) error {
						var matchers, expected []map[string]interface{}
						if err := json.Unmarshal([]byte(value), &matchers); err != nil {
							return err
						}
						if err := json.Unmarshal([]byte(expectdMatchers2), &expected); err != nil {
							return err
						}
						if cmp.Equal(matchers, expected) {
							return nil
						} else {
							return fmt.Errorf("got invalid matchers: %#v", matchers)
						}
					}),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteAlertLogMeasureRuleDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewLogMeasureRuleOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_alert_log_measure_rule" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.Attributes["alert_project_id"], uuid.MustParse(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists monitoring suite alert log measure rule: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteAlertLogMeasureRuleExists(n string, routing *v1.LogMeasureRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no alert log measure rule ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewLogMeasureRuleOp(client.MonitoringSuiteClient)
		found, err := op.Read(context.Background(), rs.Primary.Attributes["alert_project_id"], uuid.MustParse(rs.Primary.ID))
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found alert log measure rule: %s", rs.Primary.ID)
		}

		*routing = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteAlertLogMeasureRule_basic = `
resource "sakura_monitoring_suite_alert_project" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_alert_log_measure_rule" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  alert_project_id = sakura_monitoring_suite_alert_project.foobar.id
  log_storage_id = {{ .arg1 }}
  metric_storage_id = {{ .arg2 }}
  rule = {
    version = "v1"
    query = {
      matchers = jsonencode([
        {
          type = "string"
          operator = "eq"
          field = "service_name"
          value = "example"
          value_list = ["example"]
        }
      ])
    }
  }
}`

var testAccSakuraMonitoringSuiteAlertLogMeasureRule_update = `
resource "sakura_monitoring_suite_alert_project" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_alert_log_measure_rule" "foobar" {
  name = "{{ .arg0 }}-updated"
  description = "description-updated"
  alert_project_id = sakura_monitoring_suite_alert_project.foobar.id
  log_storage_id = {{ .arg1 }}
  metric_storage_id = {{ .arg2 }}
  rule = {
    version = "v1"
    query = {
      matchers = jsonencode([
        {
          type = "or"
          matchers = [
            {
              type = "string"
              field = "text_payload"
              value = "value"
              operator = "eq"
              value_list = []
            },
            {
              type = "number"
              field = "http_request_size"
              value = 1000
              operator = "gt"
              value_list = []
            }
          ]
        }
      ])
    }
  }
}`
