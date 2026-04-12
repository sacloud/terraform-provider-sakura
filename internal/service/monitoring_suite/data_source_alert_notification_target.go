// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	v1 "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type alertNotificationTargetDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &alertNotificationTargetDataSource{}
	_ datasource.DataSourceWithConfigure = &alertNotificationTargetDataSource{}
)

func NewAlertNotificationTargetDataSource() datasource.DataSource {
	return &alertNotificationTargetDataSource{}
}

func (d *alertNotificationTargetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert_notification_target"
}

func (d *alertNotificationTargetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type alertNotificationTargetDataSourceModel struct {
	alertNotificationTargetBaseModel
}

func (d *alertNotificationTargetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":               common.SchemaDataSourceId("Monitoring Suite Alert Notification Target"),
			"description":      common.SchemaDataSourceDescription("Monitoring Suite Alert Notification Target"),
			"alert_project_id": schemaDataSourceAlertProjectId(),
			"service_type": schema.StringAttribute{
				Computed:    true,
				Description: "The service type of the Alert Notification Target.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the Alert Notification Target.",
			},
			"config": schema.StringAttribute{
				Computed:    true,
				Description: "The config of the Alert Notification Target.",
			},
		},
		MarkdownDescription: "Get information about an existing Monitoring Suite Alert Notification Target.",
	}
}

func (d *alertNotificationTargetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data alertNotificationTargetDataSourceModel
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

	op := monitoringsuite.NewNotificationTargetOp(d.client)
	var alert *v1.NotificationTarget
	var err error
	alert, err = op.Read(ctx, data.AlertProjectID.ValueString(), id)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Alert Notification Target[%s]: %s", sid, err))
		return
	}

	data.updateState(alert)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
