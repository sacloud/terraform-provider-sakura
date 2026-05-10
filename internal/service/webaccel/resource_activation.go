// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package webaccel

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/webaccel-api-go"
)

type webAccelActivationResource struct {
	client *webaccel.Client
}

var (
	_ resource.Resource                = &webAccelActivationResource{}
	_ resource.ResourceWithConfigure   = &webAccelActivationResource{}
	_ resource.ResourceWithImportState = &webAccelActivationResource{}
)

func NewWebAccelActivationResource() resource.Resource {
	return &webAccelActivationResource{}
}

func (r *webAccelActivationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webaccel_activation"
}

func (r *webAccelActivationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.WebaccelClient
}

type webAccelActivationResourceModel struct {
	ID      types.String `tfsdk:"id"`
	SiteID  types.String `tfsdk:"site_id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (r *webAccelActivationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("WebAccel Activation"),
			"site_id": schema.StringAttribute{
				Required:    true,
				Description: "The site ID of WebAccel.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the WebAccel activation is enabled or not.",
			},
		},
		MarkdownDescription: "Manages a WebAccel activation.",
	}
}

func (r *webAccelActivationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *webAccelActivationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webAccelActivationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := webaccel.NewOp(r.client)
	site, err := op.Read(ctx, plan.SiteID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read WebAccel site[%s]: %s", plan.SiteID.ValueString(), err))
		return
	}

	statusString := expandWebAccelActivationStatus(plan.Enabled)
	if statusString != site.Status {
		if _, err := op.UpdateStatus(ctx, plan.SiteID.ValueString(), &webaccel.UpdateSiteStatusRequest{Status: statusString}); err != nil {
			resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to update WebAccel activation[%s]: %s", plan.SiteID.ValueString(), err))
			return
		}
	}

	plan.ID = types.StringValue(site.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webAccelActivationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webAccelActivationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := state.ID.ValueString()
	site, err := webaccel.NewOp(r.client).Read(ctx, siteID)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read WebAccel activation[%s]: %s", siteID, err))
		return
	}

	state.ID = types.StringValue(site.ID)
	state.SiteID = types.StringValue(site.ID)
	state.Enabled = types.BoolValue(site.Status == "enabled")
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *webAccelActivationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state webAccelActivationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Enabled.Equal(state.Enabled) {
		statusString := expandWebAccelActivationStatus(plan.Enabled)
		if _, err := webaccel.NewOp(r.client).UpdateStatus(ctx, plan.SiteID.ValueString(), &webaccel.UpdateSiteStatusRequest{Status: statusString}); err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update WebAccel activation[%s]: %s", plan.SiteID.ValueString(), err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webAccelActivationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webAccelActivationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := state.ID.ValueString()
	if _, err := webaccel.NewOp(r.client).UpdateStatus(ctx, siteID, &webaccel.UpdateSiteStatusRequest{Status: "disabled"}); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to disable WebAccel activation[%s]: %s", siteID, err))
		return
	}
}

func expandWebAccelActivationStatus(enabled types.Bool) string {
	if enabled.ValueBool() {
		return "enabled"
	}
	return "disabled"
}
