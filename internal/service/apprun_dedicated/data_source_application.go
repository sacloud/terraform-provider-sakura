// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	app "github.com/sacloud/apprun-dedicated-api-go/apis/application"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

type appDataSource struct{ dataSourceClient }
type appDataSourceModel struct{ appModel }

var (
	_ datasource.DataSource              = &appDataSource{}
	_ datasource.DataSourceWithConfigure = &appDataSource{}
)

func NewAppDataSource() datasource.DataSource { return &appDataSource{dataSourceNamed("application")} }

func (d *appDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	id := d.schemaID()

	name := d.schemaName()

	clusterID := d.schemaClusterID()

	clusterName := schema.StringAttribute{
		Computed:    true,
		Description: "The name of the cluster",
	}

	activeVersion := schema.Int32Attribute{
		Computed:    true,
		Description: "The active version of the application",
	}

	desiredCount := schema.Int32Attribute{
		Computed:    true,
		Description: "The desired count of the application",
	}

	scalingCooldownSeconds := schema.Int32Attribute{
		Computed:    true,
		Description: "The scaling cooldown seconds of the application",
	}

	res.Schema = schema.Schema{
		Description: "Information about an AppRun dedicated application",
		Attributes: map[string]schema.Attribute{
			"id":                       id,
			"name":                     name,
			"cluster_id":               clusterID,
			"cluster_name":             clusterName,
			"active_version":           activeVersion,
			"desired_count":            desiredCount,
			"scaling_cooldown_seconds": scalingCooldownSeconds,
		},
	}
}

func (d *appDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state appDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	var appID *v1.ApplicationID

	if state.ID.IsNull() {
		// Lookup by name
		appID = d.byName(ctx, req, res, &state)
	} else {
		// Lookup by ID
		appID = d.byID(ctx, req, res, &state)
	}

	if appID == nil {
		return
	}

	detail, err := d.api().Read(ctx, *appID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated application: %s", err))
		return
	}

	state.updateState(ctx, detail)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (d *appDataSource) byID(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse, state *appDataSourceModel) *v1.ApplicationID {
	appID, err := state.applicationID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid ID", fmt.Sprintf("failed to parse application ID: %s", err))
	}

	return &appID
}

func (d *appDataSource) byName(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse, state *appDataSourceModel) *v1.ApplicationID {
	apps, err := listed(func(c *string) ([]v1.ReadApplicationDetail, *string, error) { return d.api().List(ctx, 10, c) })

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list AppRun Dedicated applications: %s", err))
		return nil
	}

	name := state.Name.ValueString()
	for _, i := range apps {
		if i.Name == name {
			return &i.ApplicationID
		}
	}

	res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("application with name %q not found", name))
	return nil
}

func (r *appDataSource) api() *app.ApplicationOp { return app.NewApplicationOp(r.client) }
