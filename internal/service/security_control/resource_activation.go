// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/saclient-go"
	seccon "github.com/sacloud/security-control-api-go"
	v1 "github.com/sacloud/security-control-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type activationResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &activationResource{}
	_ resource.ResourceWithConfigure   = &activationResource{}
	_ resource.ResourceWithImportState = &activationResource{}
)

func NewActivationResource() resource.Resource {
	return &activationResource{}
}

func (r *activationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_control_activation"
}

func (r *activationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.SecurityControlClient
}

type activationResourceModel struct {
	activationBaseModel
	NoActionOnDelete types.Bool     `tfsdk:"no_action_on_delete"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

func (r *activationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_principal_id": schema.StringAttribute{
				Required:    true,
				Description: "The Service Principal ID associated with the Project",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the Security Control is enabled",
			},
			"automated_action_limit": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of registerable automated actions",
			},
			"no_action_on_delete": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "No action when the resource is deleted. Keep existing Activation as is.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Security Control's Activation.",
	}
}

func (r *activationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *activationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan activationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	actOp := seccon.NewActivationOp(r.client)
	_, err := actOp.Read(ctx)
	if err != nil {
		if saclient.IsNotFoundError(err) { // アクティベーションされてないアカウントではReadで404が返るためそれを利用する
			_, err := actOp.Create(ctx, plan.ServicePrincipalID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to activate Security Control: %s", err))
				return
			}
		} else {
			resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read Security Control Activation status: %s", err))
			return
		}
	}

	act, err := actOp.Update(ctx, plan.ServicePrincipalID.ValueString(), plan.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to update Security Control Activation status: %s", err))
		return
	}

	plan.updateState(act)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *activationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state activationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	act := getActivation(ctx, r.client, &resp.Diagnostics)
	if act == nil {
		return
	}

	state.updateState(act)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *activationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan activationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	actOp := seccon.NewActivationOp(r.client)
	act, err := actOp.Update(ctx, plan.ServicePrincipalID.ValueString(), plan.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Security Control's Activation status: %s", err))
		return
	}

	plan.updateState(act)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *activationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state activationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.NoActionOnDelete.ValueBool() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	actOp := seccon.NewActivationOp(r.client)
	_, err := actOp.Update(ctx, state.ServicePrincipalID.ValueString(), false)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to disable Security Control Activation: %s", err))
		return
	}
}

func getActivation(ctx context.Context, client *v1.Client, diags *diag.Diagnostics) *v1.ActivationOutput {
	actOp := seccon.NewActivationOp(client)
	act, err := actOp.Read(ctx)
	if err != nil {
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Security Control's Activation status: %s", err))
		return nil
	}

	return act
}
