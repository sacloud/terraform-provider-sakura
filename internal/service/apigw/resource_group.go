// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	"github.com/hashicorp/terraform-plugin-framework/path"

	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type apigwGroupResource struct {
	client *v1.Client
}

func NewApigwGroupResource() resource.Resource {
	return &apigwGroupResource{}
}

var (
	_ resource.Resource                = &apigwGroupResource{}
	_ resource.ResourceWithConfigure   = &apigwGroupResource{}
	_ resource.ResourceWithImportState = &apigwGroupResource{}
)

func (r *apigwGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_group"
}

func (r *apigwGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwGroupResourceModel struct {
	apigwGroupBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *apigwGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaResourceId("API Gateway Group"),
			"name":       schemaResourceAPIGWName("API Gateway Group"),
			"tags":       common.SchemaResourceTags("API Gateway Group"),
			"created_at": schemaResourceAPIGWCreatedAt("API Gateway Group"),
			"updated_at": schemaResourceAPIGWUpdatedAt("API Gateway Group"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manage an API Gateway group.",
	}
}

func (r *apigwGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *apigwGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apigwGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	groupOp := apigw.NewGroupOp(r.client)
	created, err := groupOp.Create(ctx, expandAPIGWGroupRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create API Gateway group: %s", err))
		return
	}

	group := getAPIGWGroup(ctx, r.client, created.ID.Value.String(), &resp.State, &resp.Diagnostics)
	if group == nil {
		return
	}

	plan.updateState(group)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data apigwGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := getAPIGWGroup(ctx, r.client, data.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if service == nil {
		return
	}

	data.updateState(service)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *apigwGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apigwGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	group := getAPIGWGroup(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if group == nil {
		return
	}

	groupOp := apigw.NewGroupOp(r.client)
	err := groupOp.Update(ctx, expandAPIGWGroupRequest(&plan), group.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update API Gateway group %s", err))
		return
	}

	group = getAPIGWGroup(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if group == nil {
		return
	}

	plan.updateState(group)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apigwGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apigwGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	group := getAPIGWGroup(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if group == nil {
		return
	}

	groupOp := apigw.NewGroupOp(r.client)
	err := groupOp.Delete(ctx, group.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete API Gateway group[%s]: %s", group.ID.Value.String(), err))
		return
	}
}

func getAPIGWGroup(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.Group {
	groupOp := apigw.NewGroupOp(client)
	group, err := groupOp.Read(ctx, uuid.MustParse(id))
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read APIGW group[%s]: %s", id, err))
		return nil
	}

	return group
}

func expandAPIGWGroupRequest(plan *apigwGroupResourceModel) *v1.Group {
	return &v1.Group{
		Name: v1.NewOptName(v1.Name(plan.Name.ValueString())),
		Tags: common.TsetToStrings(plan.Tags),
	}
}
