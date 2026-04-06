// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type dashboardDataSource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ datasource.DataSource              = &dashboardDataSource{}
	_ datasource.DataSourceWithConfigure = &dashboardDataSource{}
)

func NewDashboardDataSource() datasource.DataSource {
	return &dashboardDataSource{}
}

func (d *dashboardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_dashboard"
}

func (d *dashboardDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type dashboardDataSourceModel struct {
	dashboardBaseModel
}

func (d *dashboardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Monitoring Suite Dashboard"),
			"name":        common.SchemaDataSourceName("Monitoring Suite Dashboard"),
			"description": common.SchemaDataSourceDescription("Monitoring Suite Dashboard"),
			"resource_id": common.SchemaDataSourceId("Monitoring Suite Dashboard"),
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The project ID of the Dashboard.",
			},
			"created_at": common.SchemaDataSourceCreatedAt("Monitoring Suite Dashboard"),
		},
		MarkdownDescription: "Get information about an existing Monitoring Suite Dashboard.",
	}
}

func (d *dashboardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dashboardDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	name := data.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Read: Attribute Error", "either 'id' or 'name' must be specified.")
		return
	}

	op := monitoringsuite.NewDashboardOp(d.client)
	var dashboard *monitoringsuiteapi.DashboardProject
	var err error
	if id != "" {
		dashboard, err = op.Read(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Dashboard[%s]: %s", id, err))
			return
		}
	} else {
		dashboards, err := op.List(ctx, nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list Dashboard resources: %s", err))
			return
		}
		dashboard, err = filterDashboardByName(dashboards, name)
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	}

	data.updateState(dashboard)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterDashboardByName(dashboards []monitoringsuiteapi.DashboardProject, name string) (*monitoringsuiteapi.DashboardProject, error) {
	match := slices.Collect(func(yield func(monitoringsuiteapi.DashboardProject) bool) {
		for _, dashboard := range dashboards {
			if dashboard.Name.Value != name {
				continue
			}
			if !yield(dashboard) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple Dashboards found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
