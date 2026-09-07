// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package networking_suite

import (
	networkingsuite "github.com/sacloud/sacloud-sdk-go/api/networking-suite"
	v1 "github.com/sacloud/sacloud-sdk-go/api/networking-suite/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

func createClient(zone string, apiClient *common.APIClient) (*v1.Client, error) {
	client, err := apiClient.SaClient2.DupWith(saclient.WithZone(zone))
	if err != nil {
		return nil, err
	}
	networkingSuiteClient, err := networkingsuite.NewClient(client)
	if err != nil {
		return nil, err
	}
	return networkingSuiteClient, nil
}
