// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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

type metricRoutingResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &metricRoutingResource{}
	_ resource.ResourceWithConfigure   = &metricRoutingResource{}
	_ resource.ResourceWithImportState = &metricRoutingResource{}
)

func NewMetricRoutingResource() resource.Resource {
	return &metricRoutingResource{}
}

func (r *metricRoutingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_metric_routing"
}

func (r *metricRoutingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type metricRoutingResourceModel struct {
	metricRoutingBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *metricRoutingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("Monitoring Suite Metric Routing"),
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Description: "The resource ID of the target service.",
			},
			"storage_id": schema.StringAttribute{
				Required:    true,
				Description: "The resource ID of the Metric Storage.",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"publisher_code": schema.StringAttribute{
				Required:    true,
				Description: "The publisher code of the target service.",
			},
			"variant": schema.StringAttribute{
				Required:    true,
				Description: "The variant of the Metric Routing.",
			},
			"created_at": common.SchemaResourceCreatedAt("Monitoring Suite Metric Routing"),
			"updated_at": common.SchemaResourceUpdatedAt("Monitoring Suite Metric Routing"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: `Manages a Monitoring Suite Metric Routing.
If you want to get publisher_code and variant value, check publishers API: https://manual.sakura.ad.jp/api/cloud/monitoring-suite/#tag/連携サービス/operation/publishers_list
		`,
	}
}

func (r *metricRoutingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *metricRoutingResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// See log_routing comment
	if r.client == nil {
		return
	}

	var config metricRoutingResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateRoutingVariant(ctx, r.client, "metrics", config.PublisherCode.ValueString(), config.Variant.ValueString()); err != nil {
		resp.Diagnostics.AddError("Config: Attribute Error", err.Error())
	}
}

func (r *metricRoutingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan metricRoutingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewMetricsRoutingOp(r.client)
	created, err := op.Create(ctx, monitoringsuite.MetricsRoutingCreateParams{
		ResourceID:       expandOptionalString(plan.ResourceID),
		MetricsStorageID: plan.StorageID.ValueString(),
		PublisherCode:    plan.PublisherCode.ValueString(),
		Variant:          plan.Variant.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Metric Routing: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricRoutingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state metricRoutingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	routing := getMetricRouting(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if routing == nil {
		return
	}

	state.updateState(routing)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *metricRoutingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan metricRoutingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewMetricsRoutingOp(r.client)
	updated, err := op.Update(ctx, uuid.MustParse(plan.ID.ValueString()), monitoringsuite.MetricsRoutingUpdateParams{
		ResourceID:       expandOptionalString(plan.ResourceID),
		MetricsStorageID: expandOptionalString(plan.StorageID),
		PublisherCode:    expandOptionalString(plan.PublisherCode),
		Variant:          expandOptionalString(plan.Variant),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Metric Routing[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricRoutingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state metricRoutingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewMetricsRoutingOp(r.client)
	if err := op.Delete(ctx, uuid.MustParse(state.ID.ValueString())); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Metric Routing[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getMetricRouting(ctx context.Context, client *monitoringsuiteapi.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.MetricsRouting {
	op := monitoringsuite.NewMetricsRoutingOp(client)
	routing, err := op.Read(ctx, uuid.MustParse(id))
	if err != nil {
		if saclient.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Metric Routing[%s]: %s", id, err))
		return nil
	}
	return routing
}
