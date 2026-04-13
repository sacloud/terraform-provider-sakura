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
	api "github.com/sacloud/api-client-go"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type metricStorageResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &metricStorageResource{}
	_ resource.ResourceWithConfigure   = &metricStorageResource{}
	_ resource.ResourceWithImportState = &metricStorageResource{}
)

func NewMetricStorageResource() resource.Resource {
	return &metricStorageResource{}
}

func (r *metricStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_metric_storage"
}

func (r *metricStorageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type metricStorageResourceModel struct {
	metricStorageBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *metricStorageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Monitoring Suite Metric Storage"),
			"name":        common.SchemaResourceName("Monitoring Suite Metric Storage"),
			"description": common.SchemaResourceDescription("Monitoring Suite Metric Storage"),
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the project to which the Metric Storage belongs.",
			},
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the Metric Storage.",
			},
			"is_system": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "The flag to indicate whether this is a system Metric Storage.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"created_at": common.SchemaResourceCreatedAt("Monitoring Suite Metric Storage"),
			"endpoints": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The endpoints of the Metric Storage.",
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						Computed:    true,
						Description: "The address of the Metric Storage endpoint.",
					},
				},
			},
			"usage": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The usage of the Metric Storage.",
				Attributes: map[string]schema.Attribute{
					"metric_routings": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of Metric Routings.",
					},
					"alert_rules": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of Alert Rules.",
					},
					"log_measure_rules": schema.Int64Attribute{
						Computed:    true,
						Description: "The number of Log Measure Rules.",
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
		MarkdownDescription: "Manages a Monitoring Suite Metric Storage.",
	}
}

func (r *metricStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *metricStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan metricStorageResourceModel
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
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Metric Storage: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state metricStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storage := getMetricsStorage(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if storage == nil {
		return
	}

	state.updateState(storage)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *metricStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan metricStorageResourceModel
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
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Metric Storage[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state metricStorageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewMetricsStorageOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Metric Storage[%s]: %s", state.ID.ValueString(), err))
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
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Metric Storage[%s]: %s", id, err))
		return nil
	}
	return storage
}
