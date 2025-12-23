// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ssh_key

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type sshKeyResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &sshKeyResource{}
	_ resource.ResourceWithConfigure   = &sshKeyResource{}
	_ resource.ResourceWithImportState = &sshKeyResource{}
)

func NewSSHKeyResource() resource.Resource {
	return &sshKeyResource{}
}

func (r *sshKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (r *sshKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type sshKeyResourceModel struct {
	sshKeyBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *sshKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("SSHKey"),
			"name":        common.SchemaResourceName("SSHKey"),
			"description": common.SchemaResourceDescription("SSHKey"),
			"public_key": schema.StringAttribute{
				Required:    true,
				Description: "The body of the public key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fingerprint": schema.StringAttribute{
				Computed:    true,
				Description: "The fingerprint of the public key.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a SSH Keys.",
	}
}

func (r *sshKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *sshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sshKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	sshKeyOp := iaas.NewSSHKeyOp(r.client)
	key, err := sshKeyOp.Create(ctx, &iaas.SSHKeyCreateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		PublicKey:   plan.PublicKey.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create SSHKey: %s", err.Error()))
		return
	}

	plan.updateState(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sshKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := getSSHKey(ctx, r.client, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	state.updateState(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sshKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	sshKeyOp := iaas.NewSSHKeyOp(r.client)
	key, err := sshKeyOp.Read(ctx, common.ExpandSakuraCloudID(plan.ID))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read SSHKey[%s]: %s", plan.ID.ValueString(), err.Error()))
		return
	}

	_, err = sshKeyOp.Update(ctx, key.ID, &iaas.SSHKeyUpdateRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update SSHKey[%s]: %s", plan.ID.ValueString(), err.Error()))
		return
	}

	key = getSSHKey(ctx, r.client, common.ExpandSakuraCloudID(plan.ID), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	plan.updateState(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sshKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	sshKeyOp := iaas.NewSSHKeyOp(r.client)
	key := getSSHKey(ctx, r.client, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	if err := sshKeyOp.Delete(ctx, key.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete SSHKey[%s]: %s", state.ID.ValueString(), err.Error()))
		return
	}
}

func getSSHKey(ctx context.Context, client *common.APIClient, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.SSHKey {
	sshKeyOp := iaas.NewSSHKeyOp(client)
	sshKey, err := sshKeyOp.Read(ctx, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read SSHKey[%d]: %s", id, err))
		return nil
	}

	return sshKey
}
