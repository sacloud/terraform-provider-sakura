// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iam-api-go"
	"github.com/sacloud/iam-api-go/apis/user"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type userResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_user"
}

func (r *userResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.IamClient
}

type userResourceModel struct {
	userBaseModel
	PasswordWO        types.String   `tfsdk:"password_wo"`
	PasswordWOVersion types.Int32    `tfsdk:"password_wo_version"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func (r *userResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("IAM User"),
			"name":        common.SchemaResourceName("IAM User"),
			"description": common.SchemaResourceDescription("IAM User"),
			"code":        schemaResourceIAMCode("IAM User"),
			"email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The email of the IAM User",
			},
			"password_wo": schema.StringAttribute{
				Required:    true,
				WriteOnly:   true,
				Description: "Password for NoSQL appliance",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo_version")),
				},
			},
			"password_wo_version": schema.Int32Attribute{
				Optional:    true,
				Description: "The version of the password_wo field. This value must be greater than 0 when set. Increment this when changing password.",
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
					int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo")),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the IAM User",
			},
			"otp": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The OTP settings of the IAM User",
				Attributes: map[string]schema.Attribute{
					"status": schema.StringAttribute{
						Computed:    true,
						Description: "The OTP status of the IAM User",
					},
					"has_recovery_code": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the IAM User has recovery code for OTP",
					},
				},
			},
			"member": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The member information of the IAM User",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The member ID associated with the IAM User",
					},
					"code": schema.StringAttribute{
						Computed:    true,
						Description: "The member code associated with the IAM User",
					},
				},
			},
			"security_key_registered": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether a security key is registered for the IAM User",
			},
			"passwordless": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether passwordless authentication is enabled for the IAM User",
			},
			"created_at": common.SchemaResourceCreatedAt("IAM User"),
			"updated_at": common.SchemaResourceUpdatedAt("IAM User"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an IAM User.",
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	userOp := iam.NewUserOp(r.client)
	res, err := userOp.Create(ctx, expandUserCreateRequest(&plan, &config))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create IAM User: %s", err))
		return
	}

	plan.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := getUser(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if user == nil {
		return
	}

	state.updateState(user)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config, state userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	userOp := iam.NewUserOp(r.client)
	_, err := userOp.Update(ctx, utils.MustAtoI(plan.ID.ValueString()), expandUserUpdateRequest(&plan, &config, &state))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IAM User[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	// TODO: Use regsiter-email API to handle email update?

	user := getUser(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if user == nil {
		return
	}

	plan.updateState(user)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	userOp := iam.NewUserOp(r.client)
	user := getUser(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if user == nil {
		return
	}

	if err := userOp.Delete(ctx, user.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete IAM User[%d]: %s", user.ID, err))
		return
	}
}

func getUser(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.User {
	userOp := iam.NewUserOp(client)
	user, err := userOp.Read(ctx, utils.MustAtoI(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read IAM User[%s]: %s", id, err.Error()))
		return nil
	}
	return user
}

func expandUserCreateRequest(model, config *userResourceModel) user.CreateParams {
	params := user.CreateParams{
		Name:        model.Name.ValueString(),
		Code:        model.Code.ValueString(),
		Description: model.Description.ValueString(),
		Password:    config.PasswordWO.ValueString(),
	}
	if utils.IsKnown(model.Email) {
		params.Email = saclient.Ptr(model.Email.ValueString())
	}
	return params
}

func expandUserUpdateRequest(model, config, state *userResourceModel) user.UpdateParams {
	params := user.UpdateParams{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
	}
	if model.PasswordWOVersion.ValueInt32() > state.PasswordWOVersion.ValueInt32() {
		params.Password = saclient.Ptr(config.PasswordWO.ValueString())
	}
	return params
}
