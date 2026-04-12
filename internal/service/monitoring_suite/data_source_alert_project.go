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

type alertProjectDataSource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ datasource.DataSource              = &alertProjectDataSource{}
	_ datasource.DataSourceWithConfigure = &alertProjectDataSource{}
)

func NewAlertProjectDataSource() datasource.DataSource {
	return &alertProjectDataSource{}
}

func (d *alertProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert_project"
}

func (d *alertProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type alertProjectDataSourceModel struct {
	alertProjectBaseModel
}

func (d *alertProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Monitoring Suite Alert Project"),
			"name":        common.SchemaDataSourceName("Monitoring Suite Alert Project"),
			"description": common.SchemaDataSourceDescription("Monitoring Suite Alert Project"),
			"resource_id": common.SchemaDataSourceId("Monitoring Suite Alert Project"),
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the project to which the Alert Project belongs.",
			},
			"created_at": common.SchemaDataSourceCreatedAt("Monitoring Suite Alert Project"),
		},
		MarkdownDescription: "Get information about an existing Monitoring Suite Alert Project.",
	}
}

func (d *alertProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data alertProjectDataSourceModel
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

	op := monitoringsuite.NewAlertProjectOp(d.client)
	var alert *monitoringsuiteapi.AlertProject
	var err error
	if id != "" {
		alert, err = op.Read(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Alert Project[%s]: %s", id, err))
			return
		}
	} else {
		alerts, err := op.List(ctx, nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list Alert Project resources: %s", err))
			return
		}
		alert, err = filterAlertProjectByName(alerts, name)
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	}

	data.updateState(alert)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterAlertProjectByName(alerts []monitoringsuiteapi.AlertProject, name string) (*monitoringsuiteapi.AlertProject, error) {
	match := slices.Collect(func(yield func(monitoringsuiteapi.AlertProject) bool) {
		for _, alert := range alerts {
			if alert.Name.Value != name {
				continue
			}
			if !yield(alert) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple Alert Projects found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
