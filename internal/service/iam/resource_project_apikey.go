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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iam-api-go"
	"github.com/sacloud/iam-api-go/apis/projectapikey"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type projectApiKeyResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &projectApiKeyResource{}
	_ resource.ResourceWithConfigure   = &projectApiKeyResource{}
	_ resource.ResourceWithImportState = &projectApiKeyResource{}
)

func NewProjectApiKeyResource() resource.Resource {
	return &projectApiKeyResource{}
}

func (r *projectApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_project_apikey"
}

func (r *projectApiKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.IamClient
}

type projectApiKeyResourceModel struct {
	projectApiKeyBaseModel
	AccessTokenSecret types.String   `tfsdk:"access_token_secret"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func (r *projectApiKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("IAM Project API Key"),
			"name":        common.SchemaResourceName("IAM Project API Key"),
			"description": common.SchemaResourceDescription("IAM Project API Key"),
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The project ID associated with the IAM Project API Key",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"iam_roles": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "The IAM roles assigned to the IAM Project API Key",
			},
			"server_resource_id": schema.StringAttribute{
				Optional:    true,
				Description: "The server resource ID of IAM Project API Key.",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Description: "The zone of IAM Project API Key.",
			},
			"access_token": schema.StringAttribute{
				Computed:    true,
				Description: "The access token of the IAM Project API Key.",
			},
			"access_token_secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The access token secret of the IAM Project API Key.",
			},
			"created_at": common.SchemaResourceCreatedAt("IAM Project API Key"),
			"updated_at": common.SchemaResourceUpdatedAt("IAM Project API Key"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an IAM Project API Key.",
	}
}

func (r *projectApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *projectApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectApiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	paKeyOp := iam.NewProjectAPIKeyOp(r.client)
	res, err := paKeyOp.Create(ctx, expandProjectApiKeyCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create IAM Project API Key: %s", err))
		return
	}

	paKey := &v1.ProjectApiKey{
		ID:               res.ID,
		Name:             res.Name,
		Description:      res.Description,
		ProjectID:        res.ProjectID,
		AccessToken:      res.AccessToken,
		IamRoles:         res.IamRoles,
		ServerResourceID: res.ServerResourceID,
		ZoneID:           res.ZoneID,
		CreatedAt:        res.CreatedAt,
		UpdatedAt:        res.UpdatedAt,
	}
	plan.updateState(paKey)
	plan.AccessTokenSecret = types.StringValue(res.AccessTokenSecret)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectApiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	paKey := getProjectApiKey(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if paKey == nil {
		return
	}

	state.updateState(paKey)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state projectApiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	projectOp := iam.NewProjectAPIKeyOp(r.client)
	_, err := projectOp.Update(ctx, utils.MustAtoI(plan.ID.ValueString()), expandProjectApiKeyUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IAM Project[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	paKey := getProjectApiKey(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if paKey == nil {
		return
	}

	plan.updateState(paKey)
	// READで取得できないため、stateの値を引き継ぐ
	plan.AccessTokenSecret = types.StringValue(state.AccessTokenSecret.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *projectApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectApiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	paKeyOp := iam.NewProjectAPIKeyOp(r.client)
	paKey := getProjectApiKey(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if paKey == nil {
		return
	}

	if err := paKeyOp.Delete(ctx, paKey.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete IAM Project API Key[%d]: %s", paKey.ID, err))
		return
	}
}

func getProjectApiKey(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.ProjectApiKey {
	paKeyOp := iam.NewProjectAPIKeyOp(client)
	paKey, err := paKeyOp.Read(ctx, utils.MustAtoI(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read IAM Project API Key[%s]: %s", id, err.Error()))
		return nil
	}
	return paKey
}

func expandProjectApiKeyCreateRequest(model *projectApiKeyResourceModel) projectapikey.CreateParams {
	params := projectapikey.CreateParams{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		ProjectID:   utils.MustAtoI(model.ProjectID.ValueString()),
		IamRoles:    common.TlistToStrings(model.IAMRoles),
	}
	if utils.IsKnown(model.ServerResourceID) {
		params.ServerResourceID = saclient.Ptr(model.ServerResourceID.ValueString())
	}
	if utils.IsKnown(model.Zone) {
		params.Zone = saclient.Ptr(model.Zone.ValueString())
	}
	return params
}

func expandProjectApiKeyUpdateRequest(model *projectApiKeyResourceModel) projectapikey.UpdateParams {
	params := projectapikey.UpdateParams{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		IamRoles:    common.TlistToStrings(model.IAMRoles),
	}
	if utils.IsKnown(model.ServerResourceID) {
		params.ServerResourceID = saclient.Ptr(model.ServerResourceID.ValueString())
	}
	if utils.IsKnown(model.Zone) {
		params.Zone = saclient.Ptr(model.Zone.ValueString())
	}
	return params
}
