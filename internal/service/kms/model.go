// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package kms

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/kms-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type kmsBaseModel struct {
	common.SakuraBaseModel
	KeyOrigin     types.String `tfsdk:"key_origin"`
	Status        types.String `tfsdk:"status"`
	LatestVersion types.Int64  `tfsdk:"latest_version"`
	CreatedAt     types.String `tfsdk:"created_at"`
	ModifiedAt    types.String `tfsdk:"modified_at"`
}

func (m *kmsBaseModel) updateState(key *v1.Key) {
	m.UpdateBaseState(key.ID, key.Name, key.Description, key.Tags)
	m.KeyOrigin = types.StringValue(string(key.KeyOrigin))
	m.Status = types.StringValue(string(key.Status))
	m.LatestVersion = types.Int64Value(int64(key.LatestVersion.Value))
	m.CreatedAt = types.StringValue(string(key.CreatedAt))
	m.ModifiedAt = types.StringValue(string(key.ModifiedAt))
}
