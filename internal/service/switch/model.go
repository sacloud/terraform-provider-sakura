// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package sw1tch

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type switchBaseModel struct {
	common.SakuraBaseModel
	IconID    types.String `tfsdk:"icon_id"`
	BridgeID  types.String `tfsdk:"bridge_id"`
	ServerIDs types.Set    `tfsdk:"server_ids"`
	Zone      types.String `tfsdk:"zone"`
}

func (model *switchBaseModel) updateState(ctx context.Context, client *common.APIClient, sw *iaas.Switch, zone string) error {
	model.UpdateBaseState(sw.ID.String(), sw.Name, sw.Description, sw.Tags)

	model.BridgeID = types.StringValue(sw.BridgeID.String())
	model.Zone = types.StringValue(zone)

	var serverIDs []string
	if sw.ServerCount > 0 {
		swOp := iaas.NewSwitchOp(client)
		searched, err := swOp.GetServers(ctx, zone, sw.ID)
		if err != nil {
			return fmt.Errorf("could not find SakuraCloud Servers: switch[%s]", err)
		}
		for _, s := range searched.Servers {
			serverIDs = append(serverIDs, s.ID.String())
		}
	}
	model.ServerIDs = common.StringsToTset(serverIDs)

	return nil
}
