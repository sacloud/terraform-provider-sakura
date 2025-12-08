// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_mq_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/sacloud/simplemq-api-go/apis/v1/queue"
	"github.com/sacloud/terraform-provider-sakura/internal/service/simple_mq"
	"github.com/sacloud/terraform-provider-sakura/internal/test"
)

func TestAccSakuraDataSourceSimpleMQ_basic(t *testing.T) {
	resourceName := "data.sakura_simple_mq.foobar"
	rand := test.RandomName()

	var q queue.CommonServiceItem
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceSimpleMQ_byName, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSimpleMQExists("sakura_simple_mq.foobar", &q),
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "expire_seconds", "345600"),
				),
			},
			{
				Config: test.BuildConfigWithArgs(testAccSakuraDataSourceSimpleMQ_byTags, rand),
				Check: resource.ComposeTestCheckFunc(
					testCheckSakuraSimpleMQExists("sakura_simple_mq.foobar", &q),
					test.CheckSakuraDataSourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rand),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.0", "tag1"),
					resource.TestCheckResourceAttr(resourceName, "tags.1", "tag2"),
					resource.TestCheckResourceAttr(resourceName, "visibility_timeout_seconds", "30"),
					resource.TestCheckResourceAttr(resourceName, "expire_seconds", "345600"),
				),
			},
		},
	})
}

var testAccSakuraDataSourceSimpleMQ_byName = `
resource "sakura_simple_mq" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]

  visibility_timeout_seconds = 30
  # expire_seconds           = 345600
}

data "sakura_simple_mq" "foobar" {
  name = "{{ .arg0 }}"

  # NOTE: resourceを先に作らせてから参照するために依存関係を明示
  depends_on = [
    sakura_simple_mq.foobar
  ]
}`

var testAccSakuraDataSourceSimpleMQ_byTags = `
resource "sakura_simple_mq" "foobar" {
  name        = "{{ .arg0 }}"
  description = "description"
  tags        = ["tag1", "tag2"]

  visibility_timeout_seconds = 30
  # expire_seconds           = 345600
}

data "sakura_simple_mq" "foobar" {
  tags = [
    "tag1"
  ]

  # NOTE: resourceを先に作らせてから参照するために依存関係を明示
  depends_on = [
    sakura_simple_mq.foobar
  ]
}`

func TestFilterSimpleMQByNameOrTags(t *testing.T) {
	t.Parallel()

	queues := []queue.CommonServiceItem{
		{
			Status: queue.Status{
				QueueName: "test-queue1",
			},
			Tags: []string{"tag1"},
		},
		{
			Status: queue.Status{
				QueueName: "test-queue2",
			},
			Tags: []string{"tag1", "tag2"},
		},
	}

	testCases := []struct {
		name      string
		queueName string
		tags      []string
		want      *queue.CommonServiceItem
		wantErr   bool
	}{
		{
			name:      "found by name",
			queueName: "test-queue1",
			want:      &queues[0],
		},
		{
			name: "found by tags",
			tags: []string{"tag2"},
			want: &queues[1],
		},
		{
			name:      "found by name & tags",
			queueName: "test-queue2",
			tags:      []string{"tag2"},
			want:      &queues[1],
		},
		{
			name:    "found multiple",
			tags:    []string{"tag1"},
			wantErr: true,
		},
		{
			name:    "not found",
			tags:    []string{"not-exist"},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := simple_mq.FilterSimpleMQByNameOrTags(queues, tc.queueName, tc.tags)
			if tc.wantErr && err == nil {
				t.Errorf("FilterSimpleMQByNameOrTags() wants error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("FilterSimpleMQByNameOrTags() error = %v", err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("FilterSimpleMQByNameOrTags() got = %v, want %v", got, tc.want)
			}
		})
	}
}
