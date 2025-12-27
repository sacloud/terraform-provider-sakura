// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	stringplanmodifier "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"

	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuitev1 "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

// モニタリングスイート　ログストレージ
type logsStorageResource struct {
	client *monitoringsuite.Client
}

var (
	_ resource.Resource                = &logsStorageResource{}
	_ resource.ResourceWithConfigure   = &logsStorageResource{}
	_ resource.ResourceWithImportState = &logsStorageResource{}
)

// provider.goから
func NewLogsStorageResource() resource.Resource {
	return &logsStorageResource{}
}

type logsStorageResourceModel struct {
	logStorageBaseModel

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *logsStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	// タイプ名　sakura_monitoring_logs_storage
	resp.TypeName = req.ProviderTypeName + "_monitoring_logs_storage"
}

func (r *logsStorageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	// TODO
	// check internal/common/config.go ProviderData setup and match accordingly
	// following is a placeholder simply for building
	cfg, ok := req.ProviderData.(*common.APIClient) // TODO adjust type + field name
	if !ok || cfg == nil || cfg.MonitoringSuiteClient == nil {
		resp.Diagnostics.AddError(
			"Unexpected provider configuration",
			"Monitoring Suite client is not available in ProviderData. Check provider.Configure and logsStorageResource.Configure.",
		)
		return
	}

	r.client = cfg.MonitoringSuiteClient
}

func (r *logsStorageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Monitoring Suite Log Storage.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the log storage.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the log storage.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the log storage.",
			},

			"classification": schema.StringAttribute{
				Optional:    true,
				Description: "Classification for the log storage. One of \"shared\" or \"separated\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				// TODO add validator?
			},

			"expire_day": schema.Int64Attribute{
				Optional:    true,
				Description: "Retention days for this log storage.",
			},

			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "Account ID that owns this log storage.",
			},
			"is_system": schema.BoolAttribute{
				Computed:    true,
				Description: "True if this is a system log storage.",
			},

			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *logsStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// terraform import sakura_monitoring_logs_storage.foo <id>
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------- CRUD ----------

func (r *logsStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Not configured", "Monitoring Suite client is nil")
		return
	}

	var plan logsStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuitev1.NewLogsStorageOp(r.client)

	params := expandLogStorageCreateParams(&plan)

	created, err := op.Create(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating log storage", err.Error())
		return
	}

	if err := plan.logStorageBaseModel.updateState(ctx, created); err != nil {
		resp.Diagnostics.AddError("Error updating state from API response", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d", created.Id)) // TODO field name

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *logsStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Not configured", "Monitoring Suite client is nil")
		return
	}

	var state logsStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutRead(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuitev1.NewLogsStorageOp(r.client)

	ls, err := op.Read(ctx, state.ID.ValueString())
	if common.IsNotFoundError(err) { // TODO adjust helper name to common package?
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading log storage", err.Error())
		return
	}

	if err := state.logStorageBaseModel.updateState(ctx, ls); err != nil {
		resp.Diagnostics.AddError("Error updating state from API response", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *logsStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Not configured", "Monitoring Suite client is nil")
		return
	}

	var plan logsStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuitev1.NewLogsStorageOp(r.client)

	params := expandLogStorageUpdateParams(&plan)

	updated, err := op.Update(ctx, plan.ID.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating log storage", err.Error())
		return
	}

	if err := plan.logStorageBaseModel.updateState(ctx, updated); err != nil {
		resp.Diagnostics.AddError("Error updating state from API response", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *logsStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Not configured", "Monitoring Suite client is nil")
		return
	}

	var state logsStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuitev1.NewLogsStorageOp(r.client)

	err := op.Delete(ctx, state.ID.ValueString())
	if common.IsNotFoundError(err) { // TODO adjust
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error deleting log storage", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

// expandLogStorageCreateParams builds the API create params from the plan.
func expandLogStorageCreateParams(m *logsStorageResourceModel) monitoringsuite.LogStorageCreateParams {
	var description *string
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		d := m.Description.ValueString()
		// empty string is OK (clear description) but still send pointer
		description = &d
	}

	var classification *monitoringsuitev1.LogStorageCreateClassification
	if !m.Classification.IsNull() && !m.Classification.IsUnknown() {
		cStr := m.Classification.ValueString()
		if cStr != "" {
			v := monitoringsuitev1.LogStorageCreateClassification(cStr)
			classification = &v
		}
	}

	return monitoringsuite.LogStorageCreateParams{
		Name:           m.Name.ValueString(),
		Description:    description,
		IsSystem:       false,
		Classification: classification,
	}
}

func expandLogStorageUpdateParams(m *logsStorageResourceModel) monitoringsuite.LogStorageUpdateParams {
	var name *string
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		n := m.Name.ValueString()
		if n != "" {
			name = &n
		}
	}

	var description *string
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		d := m.Description.ValueString()
		// empty string is allowed so clearing description
		description = &d
	}

	var expireDay *int64
	if !m.ExpireDay.IsNull() && !m.ExpireDay.IsUnknown() {
		v := m.ExpireDay.ValueInt64()
		expireDay = &v
	}

	return monitoringsuite.LogStorageUpdateParams{
		Name:        name,
		Description: description,
		ExpireDay:   expireDay,
	}
}
