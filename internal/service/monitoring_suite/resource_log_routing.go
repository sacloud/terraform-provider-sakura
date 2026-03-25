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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type logRoutingResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &logRoutingResource{}
	_ resource.ResourceWithConfigure   = &logRoutingResource{}
	_ resource.ResourceWithImportState = &logRoutingResource{}
)

func NewLogRoutingResource() resource.Resource {
	return &logRoutingResource{}
}

func (r *logRoutingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_log_routing"
}

func (r *logRoutingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type logRoutingResourceModel struct {
	logRoutingBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *logRoutingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("Monitoring Suite log routing"),
			"resource_id": schema.StringAttribute{
				Required:    true,
				Description: "The resource ID of the log storage.",
			},
			"storage_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the log storage.",
			},
			"publisher_code": schema.StringAttribute{
				Required:    true,
				Description: "The publisher code of the log routing.",
			},
			"variant": schema.StringAttribute{
				Required:    true,
				Description: "The variant of the log routing.",
			},
			"created_at": common.SchemaResourceCreatedAt("Monitoring Suite log routing"),
			"updated_at": common.SchemaResourceUpdatedAt("Monitoring Suite log routing"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Monitoring Suite log routing.",
	}
}

func (r *logRoutingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *logRoutingResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// 最初の設定チェックではConfigureより先にValidateConfigが呼ばれるため、その時はスキップする。
	// plan/apply時ではConfiureの後にValidateConfigが呼ばれるため、その時は実行する。
	// 将来的にはterraform validate段階でこのチェックを走らせるようにしたい。
	if r.client == nil {
		return
	}

	var config logRoutingResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateRoutingVariant(ctx, r.client, config.PublisherCode.ValueString(), config.Variant.ValueString()); err != nil {
		resp.Diagnostics.AddError("Config: Attribute Error", err.Error())
	}
}

func (r *logRoutingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan logRoutingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewLogRoutingOp(r.client)
	created, err := op.Create(ctx, monitoringsuite.LogsRoutingCreateParams{
		ResourceID:    expandOptionalString(plan.ResourceID),
		LogStorageID:  plan.StorageID.ValueString(),
		PublisherCode: plan.PublisherCode.ValueString(),
		Variant:       plan.Variant.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create log routing: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *logRoutingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state logRoutingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	routing := getLogRouting(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if routing == nil {
		return
	}

	state.updateState(routing)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *logRoutingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan logRoutingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	if err := validateRoutingVariant(ctx, r.client, plan.PublisherCode.ValueString(), plan.Variant.ValueString()); err != nil {
		resp.Diagnostics.AddError("Update: Attribute Error", err.Error())
		return
	}

	op := monitoringsuite.NewLogRoutingOp(r.client)
	updated, err := op.Update(ctx, uuid.MustParse(plan.ID.ValueString()), monitoringsuite.LogsRoutingUpdateParams{
		ResourceID:    expandOptionalString(plan.ResourceID),
		LogStorageID:  expandOptionalString(plan.StorageID),
		PublisherCode: expandOptionalString(plan.PublisherCode),
		Variant:       expandOptionalString(plan.Variant),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update log routing[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *logRoutingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state logRoutingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewLogRoutingOp(r.client)
	if err := op.Delete(ctx, uuid.MustParse(state.ID.ValueString())); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete log routing[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getLogRouting(ctx context.Context, client *monitoringsuiteapi.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.LogRouting {
	op := monitoringsuite.NewLogRoutingOp(client)
	routing, err := op.Read(ctx, uuid.MustParse(id))
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read log routing[%s]: %s", id, err))
		return nil
	}
	return routing
}
