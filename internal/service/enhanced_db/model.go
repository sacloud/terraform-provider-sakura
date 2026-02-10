// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package enhanced_db

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-service-go/enhanceddb/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type enhancedDBBaseModel struct {
	common.SakuraBaseModel
	IconID          types.String `tfsdk:"icon_id"`
	DatabaseName    types.String `tfsdk:"database_name"`
	DatabaseType    types.String `tfsdk:"database_type"`
	Region          types.String `tfsdk:"region"`
	Hostname        types.String `tfsdk:"hostname"`
	AllowedNetworks types.List   `tfsdk:"allowed_networks"`
	MaxConnections  types.Int64  `tfsdk:"max_connections"`
}

func (model *enhancedDBBaseModel) updateState(edb *builder.EnhancedDB) {
	model.UpdateBaseState(edb.ID.String(), edb.Name, edb.Description, edb.Tags)
	model.DatabaseName = types.StringValue(edb.DatabaseName)
	model.DatabaseType = types.StringValue(string(edb.DatabaseType))
	model.Region = types.StringValue(string(edb.Region))
	model.Hostname = types.StringValue(edb.HostName)
	model.MaxConnections = types.Int64Value(int64(edb.Config.MaxConnections))
	if len(edb.Config.AllowedNetworks) == 0 {
		model.AllowedNetworks = types.ListNull(types.StringType)
	} else {
		model.AllowedNetworks = common.StringsToTlist(edb.Config.AllowedNetworks)
	}
	if edb.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(edb.IconID.String())
	}
}
