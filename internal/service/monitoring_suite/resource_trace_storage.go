// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
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
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type traceStorageResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &traceStorageResource{}
	_ resource.ResourceWithConfigure   = &traceStorageResource{}
	_ resource.ResourceWithImportState = &traceStorageResource{}
)

func NewTraceStorageResource() resource.Resource {
	return &traceStorageResource{}
}

func (r *traceStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_trace_storage"
}

func (r *traceStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type traceStorageResourceModel struct {
	traceStorageBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *traceStorageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Monitoring Suite Trace Storage"),
			"name":        common.SchemaResourceName("Monitoring Suite Trace Storage"),
			"description": common.SchemaResourceDescription("Monitoring Suite Trace Storage"),
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the project to which the Trace Storage belongs.",
			},
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the Trace Storage.",
			},
			"retention_period_days": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The retention period days of the Trace Storage.",
				Validators: []validator.Int32{
					int32validator.Between(1, 730),
				},
			},
			"created_at": common.SchemaResourceCreatedAt("Monitoring Suite Trace Storage"),
			"endpoints": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The endpoints of the Trace Storage.",
				Attributes: map[string]schema.Attribute{
					"ingester": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "The ingester endpoint for the Trace Storage.",
						Attributes: map[string]schema.Attribute{
							"address": schema.StringAttribute{
								Computed:    true,
								Description: "The ingester address for the Trace Storage.",
							},
							"insecure": schema.BoolAttribute{
								Computed:    true,
								Description: "The flag to indicate whether the ingester uses insecure connection.",
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
		MarkdownDescription: "Manages a Monitoring Suite Trace Storage.",
	}
}

func (r *traceStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *traceStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan traceStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewTracesStorageOp(r.client)
	created, err := op.Create(ctx, monitoringsuite.TracesStorageCreateParams{
		Name:        plan.Name.ValueString(),
		Description: expandOptionalString(plan.Description),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Trace Storage: %s", err))
		return
	}

	if utils.IsKnown(plan.RetentionPeriodDays) {
		id := utils.ItoA(created.ID)
		res, err := op.SetExpire(ctx, id, int(plan.RetentionPeriodDays.ValueInt32()))
		if err != nil {
			resp.Diagnostics.AddWarning("Create: API Error", fmt.Sprintf("failed to set retention period days for Trace Storage[%s]. Set manually via Control Panels: %s", id, err))
			return
		}
		created.RetentionPeriodDays = res.RetentionPeriodDays
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *traceStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state traceStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storage := getTraceStorage(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if storage == nil {
		return
	}

	state.updateState(storage)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *traceStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state traceStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewTracesStorageOp(r.client)
	name := plan.Name.ValueString()
	updated, err := op.Update(ctx, plan.ID.ValueString(), monitoringsuite.TracesStorageUpdateParams{
		Name:        &name,
		Description: expandOptionalString(plan.Description),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Trace Storage[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	if utils.IsKnown(plan.RetentionPeriodDays) && plan.RetentionPeriodDays.ValueInt32() != state.RetentionPeriodDays.ValueInt32() {
		id := utils.ItoA(updated.ID)
		res, err := op.SetExpire(ctx, id, int(plan.RetentionPeriodDays.ValueInt32()))
		if err != nil {
			resp.Diagnostics.AddWarning("Update: API Error", fmt.Sprintf("failed to set retention period days for Trace Storage[%s]. Set manually via Control Panels: %s", id, err))
			return
		}
		updated.RetentionPeriodDays = res.RetentionPeriodDays
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *traceStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state traceStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewTracesStorageOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Trace Storage[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getTraceStorage(ctx context.Context, client *monitoringsuiteapi.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.TraceStorage {
	op := monitoringsuite.NewTracesStorageOp(client)
	storage, err := op.Read(ctx, id)
	if err != nil {
		if saclient.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Trace Storage[%s]: %s", id, err))
		return nil
	}
	return storage
}
