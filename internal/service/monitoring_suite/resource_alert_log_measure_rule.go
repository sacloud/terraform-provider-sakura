// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	v1 "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	saclient "github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type alertLogMeasureRuleResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &alertLogMeasureRuleResource{}
	_ resource.ResourceWithConfigure   = &alertLogMeasureRuleResource{}
	_ resource.ResourceWithImportState = &alertLogMeasureRuleResource{}
)

func NewAlertLogMeasureRuleResource() resource.Resource {
	return &alertLogMeasureRuleResource{}
}

func (r *alertLogMeasureRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert_log_measure_rule"
}

func (r *alertLogMeasureRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type alertLogMeasureRuleResourceModel struct {
	alertLogMeasureRuleBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *alertLogMeasureRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":               common.SchemaResourceId("Monitoring Suite Alert Log Measure Rule"),
			"name":             common.SchemaResourceName("Monitoring Suite Alert Log Measure Rule"),
			"description":      common.SchemaResourceDescription("Monitoring Suite Alert Log Measure Rule"),
			"alert_project_id": schemaResourceAlertProjectId(),
			"log_storage_id": schema.StringAttribute{
				Required:    true,
				Description: "The resource ID of the Log Storage.",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"metric_storage_id": schema.StringAttribute{
				Required:    true,
				Description: "The resource ID of the Metric Storage.",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"rule": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The rule of the Alert Log Measure Rule.",
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Required:    true,
						Description: "The version of the rule.",
					},
					"query": schema.SingleNestedAttribute{
						Required:    true,
						Description: "The query of the rule.",
						Attributes: map[string]schema.Attribute{
							"matchers": schema.StringAttribute{
								// jsonの比較でdiffが出ないようにNormalizedを使用する
								CustomType:  jsontypes.NormalizedType{},
								Required:    true,
								Description: "The matchers of the query in JSON format. See https://manual.sakura.ad.jp/api/cloud/portal/?api=monitoring-suite-api#tag/%E3%82%A2%E3%83%A9%E3%83%BC%E3%83%88/operation/alerts_projects_log_measure_rules_create for matcher format.",
							},
						},
					},
				},
			},
			"created_at": common.SchemaResourceCreatedAt("Monitoring Suite Alert Log Measure Rule"),
			"updated_at": common.SchemaResourceUpdatedAt("Monitoring Suite Alert Log Measure Rule"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: `Manages a Monitoring Suite Alert Log Measure Rule.`,
	}
}

func (r *alertLogMeasureRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *alertLogMeasureRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertLogMeasureRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	ruleParam, err := expandAlertLogMeasureRuleRule(plan.Rule)
	if err != nil {
		resp.Diagnostics.AddError("Create: Attribute Error", fmt.Sprintf("failed to expand rule query matchers: %s", err))
		return
	}

	op := monitoringsuite.NewLogMeasureRuleOp(r.client)
	created, err := op.Create(ctx, plan.AlertProjectID.ValueString(), monitoringsuite.LogMeasureRuleCreateParams{
		Name:             expandOptionalString(plan.Name),
		Description:      expandOptionalString(plan.Description),
		LogStorageID:     plan.LogStorageID.ValueString(),
		MetricsStorageID: plan.MetricStorageID.ValueString(),
		Rule:             ruleParam,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Log Measure Rule: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertLogMeasureRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertLogMeasureRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := getLogMeasureRule(ctx, r.client, state.AlertProjectID.ValueString(), state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if rule == nil {
		return
	}

	state.updateState(rule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *alertLogMeasureRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertLogMeasureRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	ruleParam, err := expandAlertLogMeasureRuleRule(plan.Rule)
	if err != nil {
		resp.Diagnostics.AddError("Update: Attribute Error", fmt.Sprintf("failed to expand rule query matchers: %s", err))
		return
	}

	op := monitoringsuite.NewLogMeasureRuleOp(r.client)
	updated, err := op.Update(ctx, plan.AlertProjectID.ValueString(), uuid.MustParse(plan.ID.ValueString()), monitoringsuite.LogMeasureRuleUpdateParams{
		Name:             expandOptionalString(plan.Name),
		Description:      expandOptionalString(plan.Description),
		LogStorageID:     expandOptionalString(plan.LogStorageID),
		MetricsStorageID: expandOptionalString(plan.MetricStorageID),
		Rule:             &ruleParam,
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Log Measure Rule[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertLogMeasureRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertLogMeasureRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewLogMeasureRuleOp(r.client)
	if err := op.Delete(ctx, state.AlertProjectID.ValueString(), uuid.MustParse(state.ID.ValueString())); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Log Measure Rule[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getLogMeasureRule(ctx context.Context, client *v1.Client, alertProjectID, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.LogMeasureRule {
	op := monitoringsuite.NewLogMeasureRuleOp(client)
	logMeasureRule, err := op.Read(ctx, alertProjectID, uuid.MustParse(id))
	if err != nil {
		if saclient.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Log Measure Rule[%s]: %s", id, err))
		return nil
	}
	return logMeasureRule
}

func expandAlertLogMeasureRuleRule(model *alertLogMeasureRuleRuleModel) (v1.LogMeasureRuleModel, error) {
	var matchers []v1.FieldMatcher
	diags := model.Query.Matchers.Unmarshal(&matchers)
	if diags.HasError() {
		return v1.LogMeasureRuleModel{}, fmt.Errorf("failed to unmarshal rule query matchers: %s", diags.Errors()[0].Detail())
	}

	return v1.LogMeasureRuleModel{
		Version: v1.LogMeasureRuleVersionEnum(model.Version.ValueString()),
		Query: v1.LogMeasureRuleV1{
			Matchers: matchers,
		},
	}, nil
}
