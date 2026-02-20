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
	"github.com/sacloud/iam-api-go/apis/sso"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type ssoResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &ssoResource{}
	_ resource.ResourceWithConfigure   = &ssoResource{}
	_ resource.ResourceWithImportState = &ssoResource{}
)

func NewSsoResource() resource.Resource {
	return &ssoResource{}
}

func (r *ssoResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_sso"
}

func (r *ssoResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.IamClient
}

type ssoResourceModel struct {
	ssoBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *ssoResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("IAM SSO"),
			"name":        common.SchemaResourceName("IAM SSO"),
			"description": common.SchemaResourceDescription("IAM SSO"),
			"idp_entity_id": schema.StringAttribute{
				Required:    true,
				Description: "The IdP entity ID of the IAM SSO",
			},
			"idp_login_url": schema.StringAttribute{
				Required:    true,
				Description: "The IdP login URL of the IAM SSO",
			},
			"idp_logout_url": schema.StringAttribute{
				Required:    true,
				Description: "The IdP logout URL of the IAM SSO",
			},
			"idp_certificate": schema.StringAttribute{
				Required:    true,
				Description: "The IdP certificate of the IAM SSO",
			},
			"sp_entity_id": schema.StringAttribute{
				Computed:    true,
				Description: "The SP entity ID of the IAM SSO",
			},
			"sp_acs_url": schema.StringAttribute{
				Computed:    true,
				Description: "The SP ACS URL of the IAM SSO",
			},
			"assigned": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the IAM SSO is assigned",
			},
			"created_at": common.SchemaResourceCreatedAt("IAM SSO"),
			"updated_at": common.SchemaResourceUpdatedAt("IAM SSO"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an IAM SSO profile.",
	}
}

func (r *ssoResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ssoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ssoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	ssoOp := iam.NewSSOOp(r.client)
	res, err := ssoOp.Create(ctx, expandSSOCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create IAM SSO: %s", err))
		return
	}

	plan.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ssoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ssoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sso := getSSO(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sso == nil {
		return
	}

	state.updateState(sso)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ssoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ssoResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	ssoOp := iam.NewSSOOp(r.client)
	_, err := ssoOp.Update(ctx, utils.MustAtoI(plan.ID.ValueString()), expandSSOUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IAM SSO[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	sso := getSSO(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sso == nil {
		return
	}

	plan.updateState(sso)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ssoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ssoResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	ssoOp := iam.NewSSOOp(r.client)
	sso := getSSO(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sso == nil {
		return
	}

	if err := ssoOp.Delete(ctx, sso.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete IAM SSO[%d]: %s", sso.ID, err))
		return
	}
}

func getSSO(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.SSOProfile {
	ssoOp := iam.NewSSOOp(client)
	sso, err := ssoOp.Read(ctx, utils.MustAtoI(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read IAM SSO[%s]: %s", id, err.Error()))
		return nil
	}
	return sso
}

func expandSSOCreateRequest(model *ssoResourceModel) sso.CreateParams {
	params := sso.CreateParams{
		Name:           model.Name.ValueString(),
		Description:    model.Description.ValueString(),
		IdpEntityID:    model.IdpEntityID.ValueString(),
		IdpLoginURL:    model.IdpLoginURL.ValueString(),
		IdpLogoutURL:   model.IdpLogoutURL.ValueString(),
		IdpCertificate: model.IdpCertificate.ValueString(),
	}
	return params
}

func expandSSOUpdateRequest(model *ssoResourceModel) sso.UpdateParams {
	params := sso.UpdateParams{
		Name:           model.Name.ValueString(),
		Description:    model.Description.ValueString(),
		IdpEntityID:    model.IdpEntityID.ValueString(),
		IdpLoginURL:    model.IdpLoginURL.ValueString(),
		IdpLogoutURL:   model.IdpLogoutURL.ValueString(),
		IdpCertificate: model.IdpCertificate.ValueString(),
	}
	return params
}
