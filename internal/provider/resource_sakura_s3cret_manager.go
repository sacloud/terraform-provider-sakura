// Copyright 2016-2025 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sakura

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	api "github.com/sacloud/api-client-go"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
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
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.secretmanagerClient
}

// TODO: model.goに切り出してdata sourceと共通化する
type secretManagerResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Name        types.String   `tfsdk:"name"`
	KmsKeyID    types.String   `tfsdk:"kms_key_id"`
	Description types.String   `tfsdk:"description"`
	Tags        types.Set      `tfsdk:"tags"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (r *secretManagerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaResourceId("SecretManager vault"),
			"name":        schemaResourceName("SecretManager vault"),
			"description": schemaResourceDescription("SecretManager vault"),
			"tags":        schemaResourceTags("SecretManager vault"),
			"kms_key_id": schema.StringAttribute{
				Required:    true,
				Description: "KMS key ID for the SecretManager vault.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *secretManagerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *secretManagerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data secretManagerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutCreate(ctx, data.Timeouts, timeout5min)
	defer cancel()

	createReq := expandSecretManagerCreateVault(&data)
	vaultOp := sm.NewVaultOp(r.client)
	vault, err := vaultOp.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("SecretManager Create Error", err.Error())
		return
	}

	data.ID = types.StringValue(vault.ID)
	data.Name = types.StringValue(vault.Name)
	data.KmsKeyID = types.StringValue(vault.KmsKeyID)
	data.Description = types.StringValue(vault.Description.Value)
	data.Tags = stringsToTset(vault.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *secretManagerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data secretManagerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vaultOp := sm.NewVaultOp(r.client)
	vault, err := vaultOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		if api.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("SecretManager Read Error", err.Error())
		return
	}

	data.ID = types.StringValue(vault.ID)
	data.Name = types.StringValue(vault.Name)
	data.KmsKeyID = types.StringValue(vault.KmsKeyID)
	data.Description = types.StringValue(vault.Description.Value)
	data.Tags = stringsToTset(vault.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *secretManagerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data secretManagerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutUpdate(ctx, data.Timeouts, timeout5min)
	defer cancel()

	vaultOp := sm.NewVaultOp(r.client)
	vault, err := vaultOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SecretManager Update's Read Error", err.Error())
		return
	}

	updateReq := expandSecretManagerUpdateVault(&data, vault)
	_, err = vaultOp.Update(ctx, vault.ID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("SecretManager Update Error", err.Error())
		return
	}

	data.ID = types.StringValue(vault.ID)
	data.Name = types.StringValue(updateReq.Name)
	data.KmsKeyID = types.StringValue(updateReq.KmsKeyID)
	data.Description = types.StringValue(updateReq.Description.Value)
	data.Tags = stringsToTset(updateReq.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *secretManagerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data secretManagerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutDelete(ctx, data.Timeouts, timeout5min)
	defer cancel()

	vaultOp := sm.NewVaultOp(r.client)
	vault, err := vaultOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		if api.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("SecretManager Delete's Read Error", err.Error())
		return
	}

	err = vaultOp.Delete(ctx, vault.ID)
	if err != nil {
		resp.Diagnostics.AddError("SecretManager Delete Error", err.Error())
		return
	}
}

func expandSecretManagerCreateVault(d *secretManagerResourceModel) v1.CreateVault {
	return v1.CreateVault{
		Name:        d.Name.ValueString(),
		KmsKeyID:    d.KmsKeyID.ValueString(),
		Description: v1.NewOptString(d.Description.ValueString()),
		Tags:        tsetToStrings(d.Tags),
	}
}

func expandSecretManagerUpdateVault(d *secretManagerResourceModel, before *v1.Vault) v1.Vault {
	req := v1.Vault{
		Name:     d.Name.ValueString(),
		KmsKeyID: before.KmsKeyID,
	}

	if d.Tags.IsNull() {
		req.Tags = before.Tags
	} else {
		req.Tags = tsetToStrings(d.Tags)
	}
	if d.Description.IsNull() {
		req.Description = before.Description
	} else {
		req.Description = v1.NewOptString(d.Description.ValueString())
	}

	return req
}
