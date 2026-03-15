// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraMonitoringSuiteLogStorage_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_log_storage.foobar"
	rand := test.RandomName()

	var storage monitoringsuiteapi.LogStorage
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteLogStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorage_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "classification", "shared"),
					resource.TestCheckResourceAttr(resourceName, "is_system", "false"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorage_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "classification", "shared"),
					resource.TestCheckResourceAttr(resourceName, "is_system", "false"),
				),
			},
		},
	})
}

func TestAccSakuraMonitoringSuiteMetricsStorage_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_metrics_storage.foobar"
	rand := test.RandomName()

	var storage monitoringsuiteapi.MetricsStorage
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteMetricsStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricsStorage_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricsStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "is_system", "false"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricsStorage_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricsStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
					resource.TestCheckResourceAttr(resourceName, "is_system", "false"),
				),
			},
		},
	})
}

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
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorage_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteTraceStorageExists(resourceName, &storage),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description-updated"),
				),
			},
		},
	})
}

func TestAccSakuraMonitoringSuiteLogStorageAccessKey_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_log_storage_access_key.foobar"
	rand := test.RandomName()

	var key monitoringsuiteapi.LogStorageAccessKey
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteLogStorageAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorageAccessKey_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteLogStorageAccessKey_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteLogStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
				),
			},
		},
	})
}

func TestAccSakuraMonitoringSuiteMetricsStorageAccessKey_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_metrics_storage_access_key.foobar"
	rand := test.RandomName()

	var key monitoringsuiteapi.MetricsStorageAccessKey
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteMetricsStorageAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricsStorageAccessKey_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricsStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteMetricsStorageAccessKey_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteMetricsStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
				),
			},
		},
	})
}

func TestAccSakuraMonitoringSuiteTraceStorageAccessKey_basic(t *testing.T) {
	resourceName := "sakura_monitoring_suite_trace_storage_access_key.foobar"
	rand := test.RandomName()

	var key monitoringsuiteapi.TraceStorageAccessKey
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		CheckDestroy:             testCheckSakuraMonitoringSuiteTraceStorageAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorageAccessKey_basic, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteTraceStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraMonitoringSuiteTraceStorageAccessKey_update, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraMonitoringSuiteTraceStorageAccessKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", "access-key-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
				),
			},
		},
	})
}

func testCheckSakuraMonitoringSuiteLogStorageDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewLogsStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_log_storage" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists monitoring suite log storage: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteMetricsStorageDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewMetricsStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_metrics_storage" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := op.Read(context.Background(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("still exists monitoring suite metrics storage: %s", rs.Primary.ID)
		}
	}
	return nil
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

func testCheckSakuraMonitoringSuiteLogStorageExists(n string, storage *monitoringsuiteapi.LogStorage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no log storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewLogsStorageOp(client.MonitoringSuiteClient)

		found, err := op.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if fmt.Sprintf("%d", found.ID) != rs.Primary.ID {
			return fmt.Errorf("not found log storage: %s", rs.Primary.ID)
		}

		*storage = *found
		return nil
	}
}

func testCheckSakuraMonitoringSuiteMetricsStorageExists(n string, storage *monitoringsuiteapi.MetricsStorage) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no metrics storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewMetricsStorageOp(client.MonitoringSuiteClient)

		found, err := op.Read(context.Background(), rs.Primary.ID)
		if err != nil {
			return err
		}

		if fmt.Sprintf("%d", found.ID) != rs.Primary.ID {
			return fmt.Errorf("not found metrics storage: %s", rs.Primary.ID)
		}

		*storage = *found
		return nil
	}
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

func testCheckSakuraMonitoringSuiteLogStorageAccessKeyDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewLogsStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_log_storage_access_key" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		uid, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}
		storageID := rs.Primary.Attributes["storage_id"]
		if storageID == "" {
			continue
		}

		_, err = op.ReadKey(context.Background(), storageID, uid)
		if err == nil {
			return fmt.Errorf("still exists monitoring suite log storage access key: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteMetricsStorageAccessKeyDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewMetricsStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_metrics_storage_access_key" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		uid, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}
		storageID := rs.Primary.Attributes["storage_id"]
		if storageID == "" {
			continue
		}

		_, err = op.ReadKey(context.Background(), storageID, uid)
		if err == nil {
			return fmt.Errorf("still exists monitoring suite metrics storage access key: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteTraceStorageAccessKeyDestroy(s *terraform.State) error {
	client := test.AccClientGetter()
	op := monitoringsuite.NewTracesStorageOp(client.MonitoringSuiteClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_monitoring_suite_trace_storage_access_key" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		uid, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}
		storageID := rs.Primary.Attributes["storage_id"]
		if storageID == "" {
			continue
		}

		_, err = op.ReadKey(context.Background(), storageID, uid)
		if err == nil {
			return fmt.Errorf("still exists monitoring suite trace storage access key: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCheckSakuraMonitoringSuiteLogStorageAccessKeyExists(n string, key *monitoringsuiteapi.LogStorageAccessKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no log storage access key ID is set")
		}

		uid, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}
		storageID := rs.Primary.Attributes["storage_id"]
		if storageID == "" {
			return errors.New("no log storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewLogsStorageOp(client.MonitoringSuiteClient)

		found, err := op.ReadKey(context.Background(), storageID, uid)
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found log storage access key: %s", rs.Primary.ID)
		}

		*key = *found
		return nil
	}
}

func testCheckSakuraMonitoringSuiteMetricsStorageAccessKeyExists(n string, key *monitoringsuiteapi.MetricsStorageAccessKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no metrics storage access key ID is set")
		}

		uid, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}
		storageID := rs.Primary.Attributes["storage_id"]
		if storageID == "" {
			return errors.New("no metrics storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewMetricsStorageOp(client.MonitoringSuiteClient)

		found, err := op.ReadKey(context.Background(), storageID, uid)
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found metrics storage access key: %s", rs.Primary.ID)
		}

		*key = *found
		return nil
	}
}

func testCheckSakuraMonitoringSuiteTraceStorageAccessKeyExists(n string, key *monitoringsuiteapi.TraceStorageAccessKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("no trace storage access key ID is set")
		}

		uid, err := uuid.Parse(rs.Primary.ID)
		if err != nil {
			return err
		}
		storageID := rs.Primary.Attributes["storage_id"]
		if storageID == "" {
			return errors.New("no trace storage ID is set")
		}

		client := test.AccClientGetter()
		op := monitoringsuite.NewTracesStorageOp(client.MonitoringSuiteClient)

		found, err := op.ReadKey(context.Background(), storageID, uid)
		if err != nil {
			return err
		}

		if found.UID.String() != rs.Primary.ID {
			return fmt.Errorf("not found trace storage access key: %s", rs.Primary.ID)
		}

		*key = *found
		return nil
	}
}

var testAccSakuraMonitoringSuiteLogStorage_basic = `
resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  classification = "shared"
  is_system = false
}
`

var testAccSakuraMonitoringSuiteLogStorage_update = `
resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description-updated"
  classification = "shared"
  is_system = false
}
`

var testAccSakuraMonitoringSuiteMetricsStorage_basic = `
resource "sakura_monitoring_suite_metrics_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  is_system = false
}
`

var testAccSakuraMonitoringSuiteMetricsStorage_update = `
resource "sakura_monitoring_suite_metrics_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description-updated"
  is_system = false
}
`

var testAccSakuraMonitoringSuiteTraceStorage_basic = `
resource "sakura_monitoring_suite_trace_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}
`

var testAccSakuraMonitoringSuiteTraceStorage_update = `
resource "sakura_monitoring_suite_trace_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description-updated"
}
`

var testAccSakuraMonitoringSuiteLogStorageAccessKey_basic = `
resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  classification = "shared"
  is_system = false
}

resource "sakura_monitoring_suite_log_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_log_storage.foobar.id
  description = "access-key"
}
`

var testAccSakuraMonitoringSuiteLogStorageAccessKey_update = `
resource "sakura_monitoring_suite_log_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  classification = "shared"
  is_system = false
}

resource "sakura_monitoring_suite_log_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_log_storage.foobar.id
  description = "access-key-updated"
}
`

var testAccSakuraMonitoringSuiteMetricsStorageAccessKey_basic = `
resource "sakura_monitoring_suite_metrics_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  is_system = false
}

resource "sakura_monitoring_suite_metrics_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_metrics_storage.foobar.id
  description = "access-key"
}
`

var testAccSakuraMonitoringSuiteMetricsStorageAccessKey_update = `
resource "sakura_monitoring_suite_metrics_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
  is_system = false
}

resource "sakura_monitoring_suite_metrics_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_metrics_storage.foobar.id
  description = "access-key-updated"
}
`

var testAccSakuraMonitoringSuiteTraceStorageAccessKey_basic = `
resource "sakura_monitoring_suite_trace_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_trace_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_trace_storage.foobar.id
  description = "access-key"
}
`

var testAccSakuraMonitoringSuiteTraceStorageAccessKey_update = `
resource "sakura_monitoring_suite_trace_storage" "foobar" {
  name = "{{ .arg0 }}"
  description = "description"
}

resource "sakura_monitoring_suite_trace_storage_access_key" "foobar" {
  storage_id = sakura_monitoring_suite_trace_storage.foobar.id
  description = "access-key-updated"
}
`
