// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type alertRuleResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &alertRuleResource{}
	_ resource.ResourceWithConfigure   = &alertRuleResource{}
	_ resource.ResourceWithImportState = &alertRuleResource{}
)

func NewAlertRuleResource() resource.Resource {
	return &alertRuleResource{}
}

func (r *alertRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert_rule"
}

func (r *alertRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type alertRuleResourceModel struct {
	alertRuleBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *alertRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":               common.SchemaResourceId("Monitoring Suite Alert Rule"),
			"name":             common.SchemaResourceName("Monitoring Suite Alert Rule"),
			"alert_project_id": schemaResourceAlertProjectId(),
			"metric_storage_id": schema.StringAttribute{
				Required:    true,
				Description: "The resource ID of the Metric Storage.",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"query": schema.StringAttribute{
				Required:    true,
				Description: "The query of the Alert Rule.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			"format": schema.StringAttribute{
				Optional:    true,
				Description: "The format of the Alert Rule.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
				},
			},
			"template": schema.StringAttribute{
				Optional:    true,
				Description: "The template of the Alert Rule.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
				},
			},
			"enabled_warning": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to enable warning level of the Alert Rule.",
			},
			"enabled_critical": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to enable critical level of the Alert Rule.",
			},
			"threshold_warning": schema.StringAttribute{
				Optional:    true,
				Description: "The threshold of warning level of the Alert Rule.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
				},
			},
			"threshold_critical": schema.StringAttribute{
				Optional:    true,
				Description: "The threshold of critical level of the Alert Rule.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
				},
			},
			"threshold_duration_warning": schema.Int64Attribute{
				Optional:    true,
				Description: "The threshold duration (in seconds) of warning level of the Alert Rule.",
			},
			"threshold_duration_critical": schema.Int64Attribute{
				Optional:    true,
				Description: "The threshold duration (in seconds) of critical level of the Alert Rule.",
			},
			"open": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the Alert Rule is open.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: `Manages a Monitoring Suite Alert Rule.`,
	}
}

func (r *alertRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *alertRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewAlertRuleOp(r.client)
	created, err := op.Create(ctx, plan.AlertProjectID.ValueString(), monitoringsuite.AlertRuleCreateParams{
		MetricsStorageID:          plan.MetricStorageID.ValueString(),
		Name:                      expandOptionalString(plan.Name),
		Query:                     plan.Query.ValueString(),
		Format:                    expandOptionalString(plan.Format),
		Template:                  expandOptionalString(plan.Template),
		EnabledWarning:            expandOptionalBool(plan.EnabledWarning),
		EnabledCritical:           expandOptionalBool(plan.EnabledCritical),
		ThresholdWarning:          expandOptionalString(plan.ThresholdWarning),
		ThresholdCritical:         expandOptionalString(plan.ThresholdCritical),
		ThresholdDurationWarning:  expandOptionalInt64(plan.ThresholdDurationWarning),
		ThresholdDurationCritical: expandOptionalInt64(plan.ThresholdDurationCritical),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Alert Project: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertRule := getAlertRule(ctx, r.client, state.AlertProjectID.ValueString(), state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if alertRule == nil {
		return
	}

	state.updateState(alertRule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *alertRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewAlertRuleOp(r.client)
	updated, err := op.Update(ctx, plan.AlertProjectID.ValueString(), uuid.MustParse(plan.ID.ValueString()), monitoringsuite.AlertRuleUpdateParams{
		MetricsStorageID:          expandOptionalString(plan.MetricStorageID),
		Name:                      expandOptionalString(plan.Name),
		Query:                     expandOptionalString(plan.Query),
		Format:                    expandOptionalString(plan.Format),
		Template:                  expandOptionalString(plan.Template),
		EnabledWarning:            expandOptionalBool(plan.EnabledWarning),
		EnabledCritical:           expandOptionalBool(plan.EnabledCritical),
		ThresholdWarning:          expandOptionalString(plan.ThresholdWarning),
		ThresholdCritical:         expandOptionalString(plan.ThresholdCritical),
		ThresholdDurationWarning:  expandOptionalInt64(plan.ThresholdDurationWarning),
		ThresholdDurationCritical: expandOptionalInt64(plan.ThresholdDurationCritical),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Alert Rule[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewAlertRuleOp(r.client)
	if err := op.Delete(ctx, state.AlertProjectID.ValueString(), uuid.MustParse(state.ID.ValueString())); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Alert Rule[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getAlertRule(ctx context.Context, client *monitoringsuiteapi.Client, alertID, id string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.AlertRule {
	op := monitoringsuite.NewAlertRuleOp(client)
	alertRule, err := op.Read(ctx, alertID, uuid.MustParse(id))
	if err != nil {
		if saclient.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Alert Rule[%s]: %s", id, err))
		return nil
	}
	return alertRule
}
