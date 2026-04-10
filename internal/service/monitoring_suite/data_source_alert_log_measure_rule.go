// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	v1 "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type alertLogMeasureRuleDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &alertLogMeasureRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &alertLogMeasureRuleDataSource{}
)

func NewAlertLogMeasureRuleDataSource() datasource.DataSource {
	return &alertLogMeasureRuleDataSource{}
}

func (d *alertLogMeasureRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert_log_measure_rule"
}

func (d *alertLogMeasureRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.MonitoringSuiteClient
}

type alertLogMeasureRuleDataSourceModel struct {
	alertLogMeasureRuleBaseModel
}

func (d *alertLogMeasureRuleDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Monitoring Suite Alert Log Measure Rule"),
			"name":        common.SchemaDataSourceName("Monitoring Suite Alert Log Measure Rule"),
			"description": common.SchemaDataSourceDescription("Monitoring Suite Alert Log Measure Rule"),
			"alert_id":    schemaDataSourceAlertId(),
			"log_storage_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the Log Storage.",
			},
			"metric_storage_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the Metric Storage.",
			},
			"rule": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The rule of the Alert Log Measure Rule.",
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Computed:    true,
						Description: "The version of the rule.",
					},
					"query": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "The query of the rule.",
						Attributes: map[string]schema.Attribute{
							"matchers": schema.StringAttribute{
								CustomType:  jsontypes.NormalizedType{},
								Computed:    true,
								Description: "The matchers of the query in JSON format. See https://manual.sakura.ad.jp/api/cloud/portal/?api=monitoring-suite-api#tag/%E3%82%A2%E3%83%A9%E3%83%BC%E3%83%88/operation/alerts_projects_log_measure_rules_create for matcher format.",
							},
						},
					},
				},
			},
			"created_at": common.SchemaDataSourceCreatedAt("Monitoring Suite Alert Log Measure Rule"),
			"updated_at": common.SchemaDataSourceUpdatedAt("Monitoring Suite Alert Log Measure Rule"),
		},
		MarkdownDescription: "Get information about an existing Monitoring Suite Alert Log Measure Rule.",
	}
}

func (d *alertLogMeasureRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data alertLogMeasureRuleDataSourceModel
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

	op := monitoringsuite.NewLogMeasureRuleOp(d.client)
	var rule *v1.LogMeasureRule
	var err error
	if id != "" {
		rule, err = op.Read(ctx, data.AlertID.ValueString(), uuid.MustParse(id))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Log Measure Rule[%s]: %s", id, err))
			return
		}
	} else {
		rules, err := op.List(ctx, data.AlertID.ValueString(), nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list Log Measure Rule resources: %s", err))
			return
		}
		rule, err = filterLogMeasureRuleByName(rules, name)
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	}

	data.updateState(rule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterLogMeasureRuleByName(rules []v1.LogMeasureRule, name string) (*v1.LogMeasureRule, error) {
	match := slices.Collect(func(yield func(v1.LogMeasureRule) bool) {
		for _, rule := range rules {
			if rule.Name.Value != name {
				continue
			}
			if !yield(rule) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple Alert Log Measure Rules found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
