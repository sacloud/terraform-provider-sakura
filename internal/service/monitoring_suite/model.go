// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type msBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (model *msBaseModel) updateBaseState(id string, name string, desc string) {
	model.ID = types.StringValue(id)
	model.Name = types.StringValue(name)
	model.Description = types.StringValue(desc)
}

type accessKeyBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	StorageID   types.String `tfsdk:"storage_id"`
	Token       types.String `tfsdk:"token"`
	Secret      types.String `tfsdk:"secret"`
}

func (m *accessKeyBaseModel) updateState(storageID string, uid string, description string, token string, secret string) {
	m.ID = types.StringValue(uid)
	m.Description = types.StringValue(description)
	m.StorageID = types.StringValue(storageID)
	m.Token = types.StringValue(token)
	m.Secret = types.StringValue(secret)
}

type optInt64 interface {
	Get() (int64, bool)
}

func optInt64ToType(value optInt64) types.Int64 {
	if v, ok := value.Get(); ok {
		return types.Int64Value(v)
	}
	return types.Int64Null()
}

type optString interface {
	Get() (string, bool)
	Or(string) string
}

type optBool interface {
	Get() (bool, bool)
}

func optBoolToType(value optBool) types.Bool {
	if v, ok := value.Get(); ok {
		return types.BoolValue(v)
	}
	return types.BoolNull()
}
