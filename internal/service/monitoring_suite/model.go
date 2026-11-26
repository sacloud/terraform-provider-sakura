// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	monitoringsuiteapi "github.com/sacloud/sacloud-sdk-go/api/monitoring-suite/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
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

type optBool interface {
	Get() (bool, bool)
}

func optBoolToType(value optBool) types.Bool {
	if v, ok := value.Get(); ok {
		return types.BoolValue(v)
	}
	return types.BoolNull()
}

func storageIDToType(value interface{ Get() (int64, bool) }) types.String {
	if id, ok := value.Get(); ok {
		return types.StringValue(strconv.FormatInt(id, 10))
	}
	return types.StringNull()
}

func expandStorageID(value types.String, attribute string, diags *diag.Diagnostics) monitoringsuiteapi.OptNilInt64 {
	if !utils.IsKnown(value) {
		return monitoringsuiteapi.OptNilInt64{}
	}
	id, err := strconv.ParseInt(value.ValueString(), 10, 64)
	if err != nil || id <= 0 {
		diags.AddAttributeError(path.Root(attribute), "Invalid Storage ID", "The ID must be a positive 64-bit integer.")
		return monitoringsuiteapi.OptNilInt64{}
	}
	return monitoringsuiteapi.NewOptNilInt64(id)
}
