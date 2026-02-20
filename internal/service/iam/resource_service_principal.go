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
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iam-api-go"
	"github.com/sacloud/iam-api-go/apis/serviceprincipal"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type servicePrincipalResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &servicePrincipalResource{}
	_ resource.ResourceWithConfigure   = &servicePrincipalResource{}
	_ resource.ResourceWithImportState = &servicePrincipalResource{}
)

func NewServicePrincipalResource() resource.Resource {
	return &servicePrincipalResource{}
}

func (r *servicePrincipalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_service_principal"
}

func (r *servicePrincipalResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.IamClient
}

type servicePrincipalResourceModel struct {
	servicePrincipalBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *servicePrincipalResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("IAM Service Principal"),
			"name":        common.SchemaResourceName("IAM Service Principal"),
			"description": common.SchemaResourceDescription("IAM Service Principal"),
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The project ID associated with the IAM Service Principal",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"created_at": common.SchemaResourceCreatedAt("IAM Service Principal"),
			"updated_at": common.SchemaResourceUpdatedAt("IAM Service Principal"),
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an IAM Service Principal.",
	}
}

func (r *servicePrincipalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *servicePrincipalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan servicePrincipalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	spOp := iam.NewServicePrincipalOp(r.client)
	res, err := spOp.Create(ctx, serviceprincipal.CreateParams{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		ProjectID:   utils.MustAtoI(plan.ProjectID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create IAM Service Principal: %s", err))
		return
	}

	plan.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *servicePrincipalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state servicePrincipalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sp := getServicePrincipal(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sp == nil {
		return
	}

	state.updateState(sp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *servicePrincipalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan servicePrincipalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	spOp := iam.NewServicePrincipalOp(r.client)
	_, err := spOp.Update(ctx, utils.MustAtoI(plan.ID.ValueString()), serviceprincipal.UpdateParams{
		Name:        plan.Name.ValueString(),
		Description: v1.NewOptString(plan.Description.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update IAM Service Principal[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	sp := getServicePrincipal(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sp == nil {
		return
	}

	plan.updateState(sp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *servicePrincipalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state servicePrincipalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	spOp := iam.NewServicePrincipalOp(r.client)
	sp := getServicePrincipal(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if sp == nil {
		return
	}

	if err := spOp.Delete(ctx, sp.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete IAM Service Principal[%d]: %s", sp.ID, err))
		return
	}
}

func getServicePrincipal(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.ServicePrincipal {
	spOp := iam.NewServicePrincipalOp(client)
	sp, err := spOp.Read(ctx, utils.MustAtoI(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read IAM Service Principal[%s]: %s", id, err.Error()))
		return nil
	}
	return sp
}
