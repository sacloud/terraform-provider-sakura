// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type alertRuleDataSource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ datasource.DataSource              = &alertRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &alertRuleDataSource{}
)

func NewAlertRuleDataSource() datasource.DataSource {
	return &alertRuleDataSource{}
}

func (d *alertRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert_rule"
}

func (d *alertRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type alertRuleDataSourceModel struct {
	alertRuleBaseModel
}

func (d *alertRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":               common.SchemaDataSourceId("Monitoring Suite Alert Rule"),
			"name":             common.SchemaDataSourceName("Monitoring Suite Alert Rule"),
			"alert_project_id": schemaDataSourceAlertProjectId(),
			"metric_storage_id": schema.StringAttribute{
				Computed:    true,
				Description: "The metric storage ID of the Alert Rule.",
			},
			"query": schema.StringAttribute{
				Computed:    true,
				Description: "The query of the Alert Rule.",
			},
			"format": schema.StringAttribute{
				Computed:    true,
				Description: "The format of the Alert Rule.",
			},
			"template": schema.StringAttribute{
				Computed:    true,
				Description: "The template of the Alert Rule.",
			},
			"enabled_warning": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to enable warning level of the Alert Rule.",
			},
			"enabled_critical": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether to enable critical level of the Alert Rule.",
			},
			"threshold_warning": schema.StringAttribute{
				Computed:    true,
				Description: "The threshold of warning level of the Alert Rule.",
			},
			"threshold_critical": schema.StringAttribute{
				Computed:    true,
				Description: "The threshold of critical level of the Alert Rule.",
			},
			"threshold_duration_warning": schema.Int64Attribute{
				Computed:    true,
				Description: "The threshold duration (in seconds) of warning level of the Alert Rule.",
			},
			"threshold_duration_critical": schema.Int64Attribute{
				Computed:    true,
				Description: "The threshold duration (in seconds) of critical level of the Alert Rule.",
			},
			"open": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the Alert Rule is open.",
			},
		},
		MarkdownDescription: "Get information about an existing Monitoring Suite Alert Rule.",
	}
}

func (d *alertRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data alertRuleDataSourceModel
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

	op := monitoringsuite.NewAlertRuleOp(d.client)
	var alertRule *monitoringsuiteapi.AlertRule
	var err error
	if id != "" {
		alertRule, err = op.Read(ctx, data.AlertProjectID.ValueString(), uuid.MustParse(id))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Alert Rule[%s]: %s", id, err))
			return
		}
	} else {
		alertRules, err := op.List(ctx, data.AlertProjectID.ValueString(), nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list Alert Rule resources: %s", err))
			return
		}
		alertRule, err = filterAlertRuleByName(alertRules, name)
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	}

	data.updateState(alertRule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterAlertRuleByName(alertRules []monitoringsuiteapi.AlertRule, name string) (*monitoringsuiteapi.AlertRule, error) {
	match := slices.Collect(func(yield func(monitoringsuiteapi.AlertRule) bool) {
		for _, alertRule := range alertRules {
			if alertRule.Name.Value != name {
				continue
			}
			if !yield(alertRule) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple Alert Rules found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
