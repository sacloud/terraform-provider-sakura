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

type alertResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &alertResource{}
	_ resource.ResourceWithConfigure   = &alertResource{}
	_ resource.ResourceWithImportState = &alertResource{}
)

func NewAlertResource() resource.Resource {
	return &alertResource{}
}

func (r *alertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert"
}

func (r *alertResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type alertResourceModel struct {
	alertBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *alertResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Monitoring Suite Alert Project"),
			"name":        common.SchemaResourceName("Monitoring Suite Alert Project"),
			"description": common.SchemaResourceDescription("Monitoring Suite Alert Project"),
			"resource_id": common.SchemaResourceId("Monitoring Suite Alert Project"),
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The project ID of the Alert Project.",
			},
			"created_at": common.SchemaResourceCreatedAt("Monitoring Suite Alert Project"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: `Manages a Monitoring Suite Alert Project.`,
	}
}

func (r *alertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *alertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewAlertProjectOp(r.client)
	created, err := op.Create(ctx, monitoringsuite.AlertProjectCreateParams{
		Name:        plan.Name.ValueString(),
		Description: expandOptionalString(plan.Description),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Alert Project: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alertProject := getAlertProject(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if alertProject == nil {
		return
	}

	state.updateState(alertProject)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *alertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewAlertProjectOp(r.client)
	updated, err := op.Update(ctx, plan.ID.ValueString(), monitoringsuite.AlertProjectUpdateParams{
		Name:        expandOptionalString(plan.Name),
		Description: expandOptionalString(plan.Description),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Alert Project[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewAlertProjectOp(r.client)
	if err := op.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Alert Project[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getAlertProject(ctx context.Context, client *monitoringsuiteapi.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.AlertProject {
	op := monitoringsuite.NewAlertProjectOp(client)
	alertProject, err := op.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Alert Project[%s]: %s", id, err))
		return nil
	}
	return alertProject
}
