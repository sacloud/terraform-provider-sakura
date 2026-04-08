// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	app "github.com/sacloud/apprun-dedicated-api-go/apis/application"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

type appID = v1.ApplicationID

type appModel struct {
	ID                     types.String `tfsdk:"id"`
	ClusterID              types.String `tfsdk:"cluster_id"`
	Name                   types.String `tfsdk:"name"`
	ClusterName            types.String `tfsdk:"cluster_name"`
	ActiveVersion          types.Int32  `tfsdk:"active_version"`
	DesiredCount           types.Int32  `tfsdk:"desired_count"`
	ScalingCooldownSeconds types.Int32  `tfsdk:"scaling_cooldown_seconds"`
}

var applicationAttrs = attrTypes{
	"id":                       types.StringType,
	"cluster_id":               types.StringType,
	"name":                     types.StringType,
	"cluster_name":             types.StringType,
	"active_version":           types.Int32Type,
	"desired_count":            types.Int32Type,
	"scaling_cooldown_seconds": types.Int32Type,
}

func (a *appModel) updateState(d *app.ApplicationDetail) {
	a.ID = uuid2StringValue(d.ApplicationID)
	a.ClusterID = uuid2StringValue(d.ClusterID)
	a.Name = types.StringValue(d.Name)
	a.ClusterName = types.StringValue(d.ClusterName)
	a.ActiveVersion = types.Int32PointerValue(d.ActiveVersion)
	a.DesiredCount = types.Int32PointerValue(d.DesiredCount)
	a.ScalingCooldownSeconds = types.Int32Value(d.ScalingCooldownSeconds)
}

func (appModel) AttributeTypes() attrTypes        { return applicationAttrs }
func (a *appModel) appId() (appID, error)         { return intoUUID[appID](a.ID) }
func (a *appModel) clusterID() (clusterID, error) { return intoUUID[clusterID](a.ClusterID) }
