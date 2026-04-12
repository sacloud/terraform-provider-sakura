// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type alertNotificationRoutingDataSource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ datasource.DataSource              = &alertNotificationRoutingDataSource{}
	_ datasource.DataSourceWithConfigure = &alertNotificationRoutingDataSource{}
)

func NewAlertNotificationRoutingDataSource() datasource.DataSource {
	return &alertNotificationRoutingDataSource{}
}

func (d *alertNotificationRoutingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert_notification_routing"
}

func (d *alertNotificationRoutingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type alertNotificationRoutingDataSourceModel struct {
	alertNotificationRoutingBaseModel
}

func (d *alertNotificationRoutingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":               common.SchemaDataSourceId("Monitoring Suite Alert Notification Routing"),
			"alert_project_id": schemaDataSourceAlertProjectId(),
			"notification_target_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Alert Notification Target.",
			},
			"resend_interval_minutes": schema.Int32Attribute{
				Computed:    true,
				Description: "The resend interval in minutes of the Alert Notification Routing.",
			},
			"match_labels": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of match label of the Alert Notification Routing.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the match label.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value of the match label.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
					},
				},
			},
			"order": schema.Int32Attribute{
				Computed:    true,
				Description: "The order of the Alert Notification Routing.",
			},
		},
		MarkdownDescription: "Get information about an existing Monitoring Suite Alert Notification Routing.",
	}
}

func (d *alertNotificationRoutingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data alertNotificationRoutingDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sid := data.ID.ValueString()
	if sid == "" {
		resp.Diagnostics.AddError("Read: Attribute Error", "'id' must be specified.")
		return
	}
	id := uuid.MustParse(sid)

	op := monitoringsuite.NewNotificationRoutingOp(d.client)
	routing, err := op.Read(ctx, data.AlertProjectID.ValueString(), id)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Alert Notification Routing[%s]: %s", sid, err))
		return
	}

	data.updateState(routing)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
