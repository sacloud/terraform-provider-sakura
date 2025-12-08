// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package kms_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	v1 "github.com/sacloud/kms-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/service/kms"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceKMS_basic(t *testing.T) {
	resourceName := "data.sakura_kms.foobar"
	rand := test.RandomName()
	var key v1.Key
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceKMS_byName, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraKMSExists("sakura_kms.foobar", &key),
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "key_origin", "generated"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "active"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceKMS_byResourceId, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraKMSExists("sakura_kms.foobar", &key),
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "key_origin", "generated"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "active"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceKMS_byName = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
}

data "sakura_kms" "foobar" {
  name = "{{ .arg0 }}"

  depends_on = [sakura_kms.foobar]
}`

var testAccSakuraDataSourceKMS_byResourceId = `
resource "sakura_kms" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]
}

data "sakura_kms" "foobar" {
  id = sakura_kms.foobar.id

  depends_on = [sakura_kms.foobar]
}`

func TestFilterKMSByName(t *testing.T) {
	t.Parallel()

	keys := []v1.Key{
		{
			Name: "test-key1",
			Tags: []string{"tag1"},
		},
		{
			Name: "test-key2",
			Tags: []string{"tag1", "tag2"},
		},
	}

	testCases := []struct {
		name    string
		keyName string
		want    *v1.Key
		wantErr bool
	}{
		{
			name:    "found by name",
			keyName: "test-key1",
			want:    &keys[0],
		},
		{
			name:    "not found",
			keyName: "not-exist",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := kms.FilterKMSByName(keys, tc.keyName)
			if tc.wantErr && err == nil {
				t.Errorf("filterKMSByName wants error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("filterKMSByName error = %v", err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("filterKMSByName got = %v, want %v", got, tc.want)
			}
		})
	}
}
