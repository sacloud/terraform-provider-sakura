// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type accessKeyBaseModel struct {
	ID          types.String `tfsdk:"id"`
	StorageID   types.String `tfsdk:"storage_id"`
	Description types.String `tfsdk:"description"`
	Token       types.String `tfsdk:"token"`
	Secret      types.String `tfsdk:"secret"`
}

func updateAccessKeyState(model *accessKeyBaseModel, storageID string, uid string, description string, token string, secret string) {
	model.ID = types.StringValue(uid)
	model.StorageID = types.StringValue(storageID)
	model.Description = types.StringValue(description)
	model.Token = types.StringValue(token)
	model.Secret = types.StringValue(secret)
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

func stringValueOrNull(value optString) types.String {
	if v, ok := value.Get(); ok {
		return types.StringValue(v)
	}
	return types.StringNull()
}
