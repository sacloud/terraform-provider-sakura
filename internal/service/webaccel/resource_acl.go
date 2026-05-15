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

type webAccelACLResource struct {
	client *webaccel.Client
}

var (
	_ resource.Resource                = &webAccelACLResource{}
	_ resource.ResourceWithConfigure   = &webAccelACLResource{}
	_ resource.ResourceWithImportState = &webAccelACLResource{}
)

func NewWebAccelACLResource() resource.Resource {
	return &webAccelACLResource{}
}

func (r *webAccelACLResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webaccel_acl"
}

func (r *webAccelACLResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.WebaccelClient
}

type webAccelACLResourceModel struct {
	ID     types.String `tfsdk:"id"`
	SiteID types.String `tfsdk:"site_id"`
	ACL    types.String `tfsdk:"acl"`
}

func (r *webAccelACLResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("WebAccel ACL"),
			"site_id": schema.StringAttribute{
				Required:    true,
				Description: "The site ID of WebAccel.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"acl": schema.StringAttribute{
				Required:    true,
				Description: "ACL definition for the WebAccel site.",
			},
		},
		MarkdownDescription: "Manages a WebAccel ACL.",
	}
}

func (r *webAccelACLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *webAccelACLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webAccelACLResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := webaccel.NewOp(r.client).UpsertACL(ctx, plan.SiteID.ValueString(), plan.ACL.ValueString()); err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to upsert WebAccel ACL: %s", err))
		return
	}

	plan.ID = types.StringValue(plan.SiteID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webAccelACLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webAccelACLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := state.ID.ValueString()

	acl, err := webaccel.NewOp(r.client).ReadACL(ctx, siteID)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read WebAccel ACL[%s]: %s", siteID, err))
		return
	}

	if acl.ACL == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(siteID)
	state.SiteID = types.StringValue(siteID)
	state.ACL = types.StringValue(acl.ACL)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *webAccelACLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state webAccelACLResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := state.ID.ValueString()
	if !plan.ACL.Equal(state.ACL) {
		if _, err := webaccel.NewOp(r.client).UpsertACL(ctx, siteID, plan.ACL.ValueString()); err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to upsert WebAccel ACL[%s]: %s", siteID, err))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *webAccelACLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webAccelACLResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID := state.ID.ValueString()
	if err := webaccel.NewOp(r.client).DeleteACL(ctx, siteID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete WebAccel ACL[%s]: %s", siteID, err))
		return
	}
}
