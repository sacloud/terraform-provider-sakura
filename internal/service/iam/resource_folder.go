// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iam-api-go"
	"github.com/sacloud/iam-api-go/apis/folder"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type folderResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &folderResource{}
	_ resource.ResourceWithConfigure   = &folderResource{}
	_ resource.ResourceWithImportState = &folderResource{}
)

func NewFolderResource() resource.Resource {
	return &folderResource{}
}

func (r *folderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_folder"
}

func (r *folderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.IamClient
}

type folderResourceModel struct {
	folderBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *folderResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("IAM Folder"),
			"name":        common.SchemaResourceName("IAM Folder"),
			"description": common.SchemaResourceDescription("IAM Folder"),
			"parent_id": schema.StringAttribute{
				Optional:    true,
				Description: "The parent folder ID of IAM Folder.",
			},
			"created_at": common.SchemaResourceCreatedAt("IAM Folder"),
			"updated_at": common.SchemaResourceUpdatedAt("IAM Folder"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an IAM Folder.",
	}
}

func (r *folderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *folderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan folderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	folderOp := iam.NewFolderOp(r.client)
	res, err := folderOp.Create(ctx, expandFolderCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create IAM Folder: %s", err))
		return
	}

	plan.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *folderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state folderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	folder := getFolder(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if folder == nil {
		return
	}

	state.updateState(folder)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *folderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan folderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	var desc *string
	if utils.IsKnown(plan.Description) {
		desc = saclient.Ptr(plan.Description.ValueString())
	}
	folderOp := iam.NewFolderOp(r.client)
	_, err := folderOp.Update(ctx, utils.MustAtoI(plan.ID.ValueString()), plan.Name.ValueString(), desc)
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IAM Folder[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	folder := getFolder(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if folder == nil {
		return
	}

	plan.updateState(folder)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *folderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state folderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	folderOp := iam.NewFolderOp(r.client)
	folder := getFolder(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if folder == nil {
		return
	}

	if err := folderOp.Delete(ctx, folder.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete IAM Folder[%d]: %s", folder.ID, err))
		return
	}
}

func getFolder(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.Folder {
	folderOp := iam.NewFolderOp(client)
	folder, err := folderOp.Read(ctx, utils.MustAtoI(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read IAM Folder[%s]: %s", id, err.Error()))
		return nil
	}
	return folder
}

func expandFolderCreateRequest(model *folderResourceModel) folder.CreateParams {
	params := folder.CreateParams{
		Name: model.Name.ValueString(),
	}
	if utils.IsKnown(model.Description) {
		params.Description = saclient.Ptr(model.Description.ValueString())
	}
	if utils.IsKnown(model.ParentID) {
		params.ParentID = saclient.Ptr(utils.MustAtoI(model.ParentID.ValueString()))
	}
	return params
}
