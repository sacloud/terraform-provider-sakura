// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iam-api-go"
	"github.com/sacloud/iam-api-go/apis/scim"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type userProvisioningResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &userProvisioningResource{}
	_ resource.ResourceWithConfigure   = &userProvisioningResource{}
	_ resource.ResourceWithImportState = &userProvisioningResource{}
)

func NewUserProvisioningResource() resource.Resource {
	return &userProvisioningResource{}
}

func (r *userProvisioningResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_user_provisioning"
}

func (r *userProvisioningResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.IamClient
}

type userProvisioningResourceModel struct {
	userProvisioningBaseModel
	SecretToken  types.String   `tfsdk:"secret_token"`
	TokenVersion types.Int32    `tfsdk:"token_version"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

func (r *userProvisioningResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaResourceId("IAM User Provisioning"),
			"name": common.SchemaResourceName("IAM User Provisioning"),
			"base_url": schema.StringAttribute{
				Computed:    true,
				Description: "The base URL of the IAM User Provisioning SCIM endpoint.",
			},
			"secret_token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The secret token for IAM User Provisioning SCIM authentication.",
			},
			"token_version": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(1),
				Description: "The version of secret_token. Increment this to regenerate the token.",
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			"created_at": common.SchemaResourceCreatedAt("IAM User Provisioning"),
			"updated_at": common.SchemaResourceUpdatedAt("IAM User Provisioning"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an IAM User Provisioning (SCIM) configuration.",
	}
}

func (r *userProvisioningResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *userProvisioningResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userProvisioningResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	scimOp := iam.NewScimOp(r.client)
	res, err := scimOp.Create(ctx, scim.CreateParams{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create IAM User Provisioning: %s", err))
		return
	}

	plan.updateState(&v1.ScimConfigurationBase{
		ID:        res.ID,
		Name:      res.Name,
		BaseURL:   res.BaseURL,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	})
	plan.SecretToken = types.StringValue(res.SecretToken)
	plan.TokenVersion = types.Int32Value(1)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userProvisioningResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userProvisioningResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scimConf := getUserProvisioning(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if scimConf == nil {
		return
	}

	state.updateState(scimConf)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userProvisioningResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state userProvisioningResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	scimOp := iam.NewScimOp(r.client)
	if plan.Name.ValueString() != state.Name.ValueString() {
		_, err := scimOp.Update(ctx, plan.ID.ValueString(), scim.UpdateParams{
			Name: plan.Name.ValueString(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IAM User Provisioning[%s]: %s", plan.ID.ValueString(), err))
			return
		}
	}

	if plan.TokenVersion.ValueInt32() > state.TokenVersion.ValueInt32() {
		regen, err := scimOp.RegenerateToken(ctx, plan.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to regenerate IAM User Provisioning token[%s]: %s", plan.ID.ValueString(), err))
			return
		}
		plan.SecretToken = types.StringValue(regen.SecretToken.Value)
	} else {
		plan.SecretToken = state.SecretToken
	}

	scimConf := getUserProvisioning(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if scimConf == nil {
		return
	}

	plan.updateState(scimConf)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userProvisioningResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userProvisioningResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	scimOp := iam.NewScimOp(r.client)
	conf := getUserProvisioning(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if conf == nil {
		return
	}

	if err := scimOp.Delete(ctx, conf.ID.String()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete IAM User Provisioning[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getUserProvisioning(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.ScimConfigurationBase {
	scimOp := iam.NewScimOp(client)
	conf, err := scimOp.Read(ctx, id)
	if err != nil {
		if saclient.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read IAM User Provisioning[%s]: %s", id, err.Error()))
		return nil
	}
	return conf
}
