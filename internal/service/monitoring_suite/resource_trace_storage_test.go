// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	monitoringsuite "github.com/sacloud/sacloud-sdk-go/api/monitoring-suite"
	monitoringsuiteapi "github.com/sacloud/sacloud-sdk-go/api/monitoring-suite/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteTraceStorage_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_trace_storage.foobar"
	rand := test.RandomName()

	var storage monitoringsuiteapi.TraceStorage
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteTraceStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorage_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteTraceStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "retention_period_days", "80"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.ingester.address"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorage_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteTraceStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "retention_period_days", "55"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoints.ingester.address"),
				),
			},
		},
	})
}

func TestAccSakuraMonitoringSuiteTraceStorage_dedicated(t *testing.T) {
	test.SkipIfEnvIsNotSet(t, "SAKURA_MONITORING_SUITE_ENABLE_DEDICATED_TEST",
		"SAKURA_MONITORING_SUITE_KMS_KEY_ID", "SAKURA_MONITORING_SUITE_SERVICE_PRINCIPAL_ID")

	resourceName := "sakura_monitoring_suite_trace_storage.foobar"
	dataSourceName := "data.sakura_monitoring_suite_trace_storage.foobar"
	rand := test.RandomName()
	kmsKeyID := os.Getenv("SAKURA_MONITORING_SUITE_KMS_KEY_ID")
	servicePrincipalID := os.Getenv("SAKURA_MONITORING_SUITE_SERVICE_PRINCIPAL_ID")

	var storage monitoringsuiteapi.TraceStorage
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteTraceStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorage_dedicated, rand, kmsKeyID, servicePrincipalID),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteTraceStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "classification", "dedicated"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", kmsKeyID),
					resource.TestCheckResourceAttr(resourceName, "service_principal_id", servicePrincipalID),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "classification", "dedicated"),
					resource.TestCheckResourceAttr(dataSourceName, "kms_key_id", kmsKeyID),
					resource.TestCheckResourceAttr(dataSourceName, "service_principal_id", servicePrincipalID),
				),
			},
		},
	})
}

func TestAccImportSakuraMonitoringSuiteTraceStorage_basic(t *testing.T) {
	rand := test.RandomName()

	checkFn := func(s []*terraform.InstanceState) error {
		if len(s) != 1 {
			return fmt.Errorf("expected 1 state: %#v", s)
		}
		expects := map[string]string{
			"name":                  rand,
			"description":           "description",
			"retention_period_days": "80",
		}

		if err := test.CompareStateMulti(s[0], expects); err != nil {
			return err
		}
		return test.StateNotEmptyMulti(s[0],
			"created_at",
			"project_id",
			"resource_id",
			"endpoints.ingester.address",
		)
	}

	resourceName := "sakura_monitoring_suite_trace_storage.foobar"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testCheckSakuraMonitoringSuiteTraceStorageDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorage_basic, rand),
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

func testCheckSakuraMonitoringSuiteTraceStorageDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewTracesStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_trace_storage" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists monitoring suite trace storage: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteTraceStorageExists(n string, storage *monitoringsuiteapi.TraceStorage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no trace storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewTracesStorageOp(client.MonitoringSuiteClient)

		found, err := op.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if fmt.Sprintf("%d", found.ID) != rs.Primary.ID {
			return fmt.Errorf("not found trace storage: %s", rs.Primary.ID)
		}

		*storage = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteTraceStorage_basic = `
resource "sakura_monitoring_suite_trace_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  retention_period_days = 80
}
`

var testAccSakuraMonitoringSuiteTraceStorage_update = `
resource "sakura_monitoring_suite_trace_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description-updated"
  retention_period_days = 55
}
`

var testAccSakuraMonitoringSuiteTraceStorage_dedicated = `
resource "sakura_monitoring_suite_trace_storage" "foobar" {
	name = "{{ .arg0 }}"
	classification = "dedicated"
	kms_key_id = "{{ .arg1 }}"
	service_principal_id = "{{ .arg2 }}"
}

data "sakura_monitoring_suite_trace_storage" "foobar" {
	id = sakura_monitoring_suite_trace_storage.foobar.id
}
`
