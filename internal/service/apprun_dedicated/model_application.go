// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	app "github.com/sacloud/apprun-dedicated-api-go/apis/application"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

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

func (a *appModel) updateState(ctx context.Context, d *app.ApplicationDetail) (ret diag.Diagnostics) {
	a.ID = uuid2StringValue(d.ApplicationID)
	a.ClusterID = uuid2StringValue(d.ClusterID)
	a.Name = types.StringValue(d.Name)
	a.ClusterName = types.StringValue(d.ClusterName)
	a.ActiveVersion = types.Int32PointerValue(d.ActiveVersion)
	a.DesiredCount = types.Int32PointerValue(d.DesiredCount)
	a.ScalingCooldownSeconds = types.Int32Value(d.ScalingCooldownSeconds)

	return
}

func (appModel) AttributeTypes() attrTypes                   { return applicationAttrs }
func (a *appModel) applicationID() (v1.ApplicationID, error) { return intoUUID[v1.ApplicationID](a.ID) }
func (a *appModel) clusterID() (v1.ClusterID, error)         { return intoUUID[v1.ClusterID](a.ClusterID) }
