// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type logStorageResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &logStorageResource{}
	_ resource.ResourceWithConfigure   = &logStorageResource{}
	_ resource.ResourceWithImportState = &logStorageResource{}
)

func NewLogStorageResource() resource.Resource {
	return &logStorageResource{}
}

func (r *logStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_log_storage"
}

func (r *logStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type logStorageResourceModel struct {
	logStorageBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *logStorageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Monitoring Suite Log Storage"),
			"name":        common.SchemaResourceName("Monitoring Suite Log Storage"),
			"description": common.SchemaResourceDescription("Monitoring Suite Log Storage"),
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the project to which the Log Storage belongs.",
			},
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the Log Storage.",
			},
			"is_system": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "The flag to indicate whether this is a system Log Storage.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"classification": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(monitoringsuiteapi.LogStorageCreateRequestClassificationShared)),
				Description: "The bucket classification of the Log Storage.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(monitoringsuiteapi.LogStorageCreateRequestClassificationShared),
						string(monitoringsuiteapi.LogStorageCreateRequestClassificationDedicated),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"expire_day": schema.Int64Attribute{
				Computed:    true,
				Description: "The expiration day of the Log Storage.",
			},
			"created_at": common.SchemaResourceCreatedAt("Monitoring Suite Log Storage"),
			"endpoints": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The endpoints of the Log Storage.",
				Attributes: map[string]schema.Attribute{
					"ingester": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "The ingester endpoint for the Log Storage.",
						Attributes: map[string]schema.Attribute{
							"address": schema.StringAttribute{
								Computed:    true,
								Description: "The ingester address for the Log Storage.",
							},
							"insecure": schema.BoolAttribute{
								Computed:    true,
								Description: "The flag to indicate whether the ingester uses insecure connection.",
							},
						},
					},
				},
			},
			"usage": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The usage of the Log Storage.",
				Attributes: map[string]schema.Attribute{
					"log_routings": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of Log Routings.",
					},
					"log_measure_rules": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of Log Measure Rules.",
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
		MarkdownDescription: "Manages a Monitoring Suite Log Storage.",
	}
}

func (r *logStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *logStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan logStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewLogsStorageOp(r.client)
	classification := monitoringsuiteapi.LogStorageCreateRequestClassification(plan.Classification.ValueString())
	created, err := op.Create(ctx, monitoringsuite.LogStorageCreateParams{
		Name:           plan.Name.ValueString(),
		Description:    expandOptionalString(plan.Description),
		IsSystem:       plan.IsSystem.ValueBool(),
		Classification: &classification,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Log Storage: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *logStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state logStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storage := getLogStorage(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if storage == nil {
		return
	}

	state.updateState(storage)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *logStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan logStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewLogsStorageOp(r.client)
	name := plan.Name.ValueString()
	updated, err := op.Update(ctx, plan.ID.ValueString(), monitoringsuite.LogStorageUpdateParams{
		Name:        &name,
		Description: expandOptionalString(plan.Description),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Log Storage[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *logStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state logStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewLogsStorageOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Log Storage[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getLogStorage(ctx context.Context, client *monitoringsuiteapi.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.LogStorage {
	op := monitoringsuite.NewLogsStorageOp(client)
	storage, err := op.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Log Storage[%s]: %s", id, err))
		return nil
	}
	return storage
}
