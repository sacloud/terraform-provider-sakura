// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	api "github.com/sacloud/api-client-go"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type metricsStorageResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &metricsStorageResource{}
	_ resource.ResourceWithConfigure   = &metricsStorageResource{}
	_ resource.ResourceWithImportState = &metricsStorageResource{}
)

func NewMetricsStorageResource() resource.Resource {
	return &metricsStorageResource{}
}

func (r *metricsStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_metrics_storage"
}

func (r *metricsStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type metricsStorageResourceModel struct {
	metricsStorageBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *metricsStorageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("Monitoring Suite metrics storage"),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the metrics storage.",
			},
			"description": common.SchemaResourceDescription("Monitoring Suite metrics storage"),
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The tags of the metrics storage.",
			},
			"icon_id": schema.StringAttribute{
				Computed:    true,
				Description: "The icon ID of the metrics storage.",
			},
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The account ID of the metrics storage.",
			},
			"resource_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The resource ID of the metrics storage.",
			},
			"is_system": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "The flag to indicate whether this is a system metrics storage.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation timestamp of the metrics storage.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The update timestamp of the metrics storage.",
			},
			"endpoints": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The endpoints of the metrics storage.",
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						Computed:    true,
						Description: "The address of the metrics storage endpoint.",
					},
				},
			},
			"usage": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The usage of the metrics storage.",
				Attributes: map[string]schema.Attribute{
					"metrics_routings": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of metrics routings.",
					},
					"alert_rules": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of alert rules.",
					},
					"log_measure_rules": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of log measure rules.",
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
		MarkdownDescription: "Manages a Monitoring Suite metrics storage.",
	}
}

func (r *metricsStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *metricsStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan metricsStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewMetricsStorageOp(r.client)
	created, err := op.Create(ctx, monitoringsuite.MetricsStorageCreateParams{
		Name:        plan.Name.ValueString(),
		Description: expandOptionalString(plan.Description),
		IsSystem:    plan.IsSystem.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create metrics storage: %s", err))
		return
	}

	updateMetricsStorageState(&plan.metricsStorageBaseModel, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricsStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state metricsStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storage := getMetricsStorage(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if storage == nil {
		return
	}

	updateMetricsStorageState(&state.metricsStorageBaseModel, storage)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *metricsStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan metricsStorageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewMetricsStorageOp(r.client)
	name := plan.Name.ValueString()
	updated, err := op.Update(ctx, plan.ID.ValueString(), monitoringsuite.MetricsStorageUpdateParams{
		Name:        &name,
		Description: expandOptionalString(plan.Description),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update metrics storage[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	updateMetricsStorageState(&plan.metricsStorageBaseModel, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricsStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state metricsStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewMetricsStorageOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete metrics storage[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getMetricsStorage(ctx context.Context, client *monitoringsuiteapi.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.MetricsStorage {
	op := monitoringsuite.NewMetricsStorageOp(client)
	storage, err := op.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read metrics storage[%s]: %s", id, err))
		return nil
	}
	return storage
}
