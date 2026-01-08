// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package dedicated_storage

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/dedicated-storage-api-go/apis/v1"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type dedicatedStorageBaseModel struct {
	common.SakuraBaseModel
	IconID types.String `tfsdk:"icon_id"`
}

func (m *dedicatedStorageBaseModel) updateState(data *v1.DedicatedStorageContract) {
	id := iaastypes.ID(data.ID)
	m.UpdateBaseState(id.String(), data.Name, data.Description, data.Tags)
	if data.Icon.IsNull() || data.Icon.Value.ID == 0 {
		m.IconID = types.StringNull()
	} else {
		m.IconID = types.StringValue(iaastypes.ID(data.Icon.Value.ID).String())
	}
}
