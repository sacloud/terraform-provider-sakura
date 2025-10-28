// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package secret_manager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type secretManagerResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &secretManagerResource{}
	_ resource.ResourceWithConfigure   = &secretManagerResource{}
	_ resource.ResourceWithImportState = &secretManagerResource{}
)

func NewSecretManagerResource() resource.Resource {
	return &secretManagerResource{}
}

func (r *secretManagerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_manager"
}

func (r *secretManagerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.SecretManagerClient
}

type secretManagerResourceModel struct {
	secretManagerBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *secretManagerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("SecretManager vault"),
			"name":        common.SchemaResourceName("SecretManager vault"),
			"description": common.SchemaResourceDescription("SecretManager vault"),
			"tags":        common.SchemaResourceTags("SecretManager vault"),
			"kms_key_id": schema.StringAttribute{
				Required:    true,
				Description: "KMS key ID for the SecretManager vault.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a Secret Manager.",
	}
}

func (r *secretManagerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *secretManagerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secretManagerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	vaultOp := sm.NewVaultOp(r.client)
	createdVault, err := vaultOp.Create(ctx, expandSecretManagerCreateVault(&plan))
	if err != nil {
		resp.Diagnostics.AddError("SecretManager Create Error", err.Error())
		return
	}

	plan.updateState(&v1.Vault{
		ID:          createdVault.ID,
		Name:        createdVault.Name,
		Description: v1.NewOptString(createdVault.Description.Value),
		Tags:        createdVault.Tags,
		KmsKeyID:    createdVault.KmsKeyID,
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secretManagerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state secretManagerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vault := getSecretManagerVault(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if vault == nil {
		return
	}

	state.updateState(vault)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *secretManagerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan secretManagerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	vault := getSecretManagerVault(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if vault == nil {
		return
	}

	vaultOp := sm.NewVaultOp(r.client)
	_, err := vaultOp.Update(ctx, vault.ID, expandSecretManagerUpdateVault(&plan, vault))
	if err != nil {
		resp.Diagnostics.AddError("SecretManager Update Error", err.Error())
		return
	}

	vault = getSecretManagerVault(ctx, r.client, vault.ID, &resp.State, &resp.Diagnostics)
	if vault == nil {
		return
	}

	plan.updateState(vault)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *secretManagerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secretManagerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	vault := getSecretManagerVault(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if vault == nil {
		return
	}

	vaultOp := sm.NewVaultOp(r.client)
	err := vaultOp.Delete(ctx, vault.ID)
	if err != nil {
		resp.Diagnostics.AddError("SecretManager Delete Error", err.Error())
		return
	}
}

func getSecretManagerVault(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diag *diag.Diagnostics) *v1.Vault {
	vaultOp := sm.NewVaultOp(client)
	vault, err := vaultOp.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diag.AddError("Get SecretManager Vault Error", err.Error())
		return nil
	}

	return vault
}

func expandSecretManagerCreateVault(model *secretManagerResourceModel) v1.CreateVault {
	return v1.CreateVault{
		Name:        model.Name.ValueString(),
		KmsKeyID:    model.KmsKeyID.ValueString(),
		Description: v1.NewOptString(model.Description.ValueString()),
		Tags:        common.TsetToStrings(model.Tags),
	}
}

func expandSecretManagerUpdateVault(model *secretManagerResourceModel, before *v1.Vault) v1.Vault {
	req := v1.Vault{
		Name:     model.Name.ValueString(),
		KmsKeyID: before.KmsKeyID,
	}

	if model.Tags.IsNull() {
		req.Tags = before.Tags
	} else {
		req.Tags = common.TsetToStrings(model.Tags)
	}
	if model.Description.IsNull() {
		req.Description = before.Description
	} else {
		req.Description = v1.NewOptString(model.Description.ValueString())
	}

	return req
}
