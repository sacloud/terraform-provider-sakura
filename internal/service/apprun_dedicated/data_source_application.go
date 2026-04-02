// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	res.Schema = schema.Schema{
		Description: "Information about an AppRun dedicated application",
		Attributes: map[string]schema.Attribute{
			"id":         d.schemaID(),
			"name":       d.schemaName(),
			"cluster_id": d.schemaClusterID(),
			"cluster_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the cluster",
			},
			"active_version": schema.Int32Attribute{
				Computed:    true,
				Description: "The active version of the application",
			},
			"desired_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The desired count of the application",
			},
			"scaling_cooldown_seconds": schema.Int32Attribute{
				Computed:    true,
				Description: "The scaling cooldown seconds of the application",
			},
		},
	}
}

func (d *appDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state appDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	var appID *appID
	var ds diag.Diagnostics

	if state.ID.IsNull() {
		// Lookup by name
		appID, ds = state.byName(ctx, d)
	} else {
		// Lookup by ID
		appID, ds = state.byId(ctx, d)
	}

	res.Diagnostics.Append(ds...)

	if appID == nil {
		return
	}

	detail, err := d.api().Read(ctx, *appID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated application: %s", err))
		return
	}

	state.updateState(detail)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (state *appDataSourceModel) byId(context.Context, *appDataSource) (ret *appID, d diag.Diagnostics) {
	appID, err := state.appId()

	if err != nil {
		d.AddError("Read: Invalid ID", fmt.Sprintf("failed to parse application ID: %s", err))
		return
	}

	ret = &appID
	return
}

func (state *appDataSourceModel) byName(ctx context.Context, d *appDataSource) (ret *appID, ds diag.Diagnostics) {
	apps, err := listed(func(c *string) ([]v1.ReadApplicationDetail, *string, error) { return d.api().List(ctx, 10, c) })

	if err != nil {
		ds.AddError("Read: API Error", fmt.Sprintf("failed to list AppRun Dedicated applications: %s", err))
		return
	}

	name := state.Name.ValueString()
	for _, i := range apps {
		if i.Name == name {
			ret = &i.ApplicationID
			return
		}
	}

	ds.AddError("Read: API Error", fmt.Sprintf("application with name %q not found", name))
	return
}

func (r *appDataSource) api() *app.ApplicationOp { return app.NewApplicationOp(r.client) }
