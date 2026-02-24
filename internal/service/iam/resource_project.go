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
	"github.com/sacloud/iam-api-go/apis/project"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type projectResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_project"
}

func (r *projectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.IamClient
}

type projectResourceModel struct {
	projectBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *projectResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("IAM Project"),
			"name":        common.SchemaResourceName("IAM Project"),
			"code":        schemaResourceIAMCode("IAM Project"),
			"description": common.SchemaResourceDescription("IAM Project"),
			"parent_folder_id": schema.StringAttribute{
				Optional:    true,
				Description: "The parent folder ID of IAM Project.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of IAM Project.",
			},
			"created_at": common.SchemaResourceCreatedAt("IAM Project"),
			"updated_at": common.SchemaResourceUpdatedAt("IAM Project"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an IAM Project.",
	}
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	projectOp := iam.NewProjectOp(r.client)
	res, err := projectOp.Create(ctx, expandProjectCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create IAM Project: %s", err))
		return
	}

	plan.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := getProject(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if project == nil {
		return
	}

	state.updateState(project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	projectOp := iam.NewProjectOp(r.client)
	_, err := projectOp.Update(ctx, utils.MustAtoI(plan.ID.ValueString()), plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IAM Project[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	project := getProject(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if project == nil {
		return
	}

	plan.updateState(project)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	projectOp := iam.NewProjectOp(r.client)
	project := getProject(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if project == nil {
		return
	}

	if err := projectOp.Delete(ctx, project.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete IAM Project[%d]: %s", project.ID, err))
		return
	}
}

func getProject(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.Project {
	projectOp := iam.NewProjectOp(client)
	project, err := projectOp.Read(ctx, utils.MustAtoI(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read IAM Project[%s]: %s", id, err.Error()))
		return nil
	}
	return project
}

func expandProjectCreateRequest(model *projectResourceModel) project.CreateParams {
	params := project.CreateParams{
		Name:        model.Name.ValueString(),
		Code:        model.Code.ValueString(),
		Description: model.Description.ValueString(),
	}
	if utils.IsKnown(model.ParentFolderID) {
		params.ParentFolderID = saclient.Ptr(utils.MustAtoI(model.ParentFolderID.ValueString()))
	}
	return params
}
