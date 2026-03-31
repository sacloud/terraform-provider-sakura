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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type dashboardResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &dashboardResource{}
	_ resource.ResourceWithConfigure   = &dashboardResource{}
	_ resource.ResourceWithImportState = &dashboardResource{}
)

func NewDashboardResource() resource.Resource {
	return &dashboardResource{}
}

func (r *dashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_dashboard"
}

func (r *dashboardResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type dashboardResourceModel struct {
	dashboardBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *dashboardResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Monitoring Suite Dashboard"),
			"name":        common.SchemaResourceName("Monitoring Suite Dashboard"),
			"description": common.SchemaResourceDescription("Monitoring Suite Dashboard"),
			"resource_id": common.SchemaResourceId("Monitoring Suite Dashboard"),
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The account ID of the Dashboard.",
			},
			"created_at": common.SchemaResourceCreatedAt("Monitoring Suite Dashboard"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: `Manages a Monitoring Suite Dashboard.`,
	}
}

func (r *dashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *dashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewDashboardOp(r.client)
	created, err := op.Create(ctx, monitoringsuite.DashboardProjectCreateParams{
		Name:        plan.Name.ValueString(),
		Description: expandOptionalString(plan.Description),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Dashboard: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboard := getDashboard(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if dashboard == nil {
		return
	}

	state.updateState(dashboard)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dashboardResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewDashboardOp(r.client)
	updated, err := op.Update(ctx, plan.ID.ValueString(), monitoringsuite.DashboardProjectUpdateParams{
		Name:        expandOptionalString(plan.Name),
		Description: expandOptionalString(plan.Description),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Dashboard[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dashboardResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewDashboardOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Dashboard[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getDashboard(ctx context.Context, client *monitoringsuiteapi.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.DashboardProject {
	op := monitoringsuite.NewDashboardOp(client)
	dashboard, err := op.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Dashboard[%s]: %s", id, err))
		return nil
	}
	return dashboard
}
