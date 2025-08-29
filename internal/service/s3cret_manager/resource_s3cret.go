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

package secret_manager

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
)

type secretManagerSecretResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &secretManagerSecretResource{}
	_ resource.ResourceWithConfigure   = &secretManagerSecretResource{}
	_ resource.ResourceWithImportState = &secretManagerSecretResource{}
)

func NewSecretManagerSecretResource() resource.Resource {
	return &secretManagerSecretResource{}
}

func (r *secretManagerSecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_manager_secret"
}

func (r *secretManagerSecretResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.SecretManagerClient
}

// TODO: model.goに切り出してdata sourceと共通化する
type secretManagerSecretResourceModel struct {
	Name     types.String   `tfsdk:"name"`
	VaultID  types.String   `tfsdk:"vault_id"`
	Version  types.Int64    `tfsdk:"version"`
	Value    types.String   `tfsdk:"value"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *secretManagerSecretResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": common.SchemaResourceName("Secret Manager's secret"),
			"vault_id": schema.StringAttribute{
				Required:    true,
				Description: "The Secret Manager's vault id.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "Version of secret value. This value is incremented by create/update.",
			},
			"value": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Secret value.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *secretManagerSecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughWithIdentity(ctx, path.Root("name"), path.Root("name"), req, resp)
}

func (r *secretManagerSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data secretManagerSecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, data.Timeouts, common.Timeout5min)
	defer cancel()

	secretOp := sm.NewSecretOp(r.client, data.VaultID.ValueString())
	createdSec, err := secretOp.Create(ctx, v1.CreateSecret{
		Name:  data.Name.ValueString(),
		Value: data.Value.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("SecretManagerSecret Create Error", err.Error())
		return
	}

	data.Name = types.StringValue(createdSec.Name)
	data.Version = types.Int64Value(int64(createdSec.LatestVersion))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *secretManagerSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data secretManagerSecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	secretOp := sm.NewSecretOp(r.client, data.VaultID.ValueString())
	secret, err := FilterSecretManagerSecretByName(ctx, secretOp, data.Name.ValueString())
	if err != nil {
		if err == common.ErrFilterNoResult {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("SecretManagerSecret Read Error", err.Error())
		return
	}

	data.Name = types.StringValue(secret.Name)
	data.Version = types.Int64Value(int64(secret.LatestVersion))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *secretManagerSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: This is same as Create, consider refactoring
	var data secretManagerSecretResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, data.Timeouts, common.Timeout5min)
	defer cancel()

	secretOp := sm.NewSecretOp(r.client, data.VaultID.ValueString())
	createdSec, err := secretOp.Create(ctx, v1.CreateSecret{
		Name:  data.Name.ValueString(),
		Value: data.Value.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("SecretManagerSecret Create Error", err.Error())
		return
	}

	data.Name = types.StringValue(createdSec.Name)
	data.Version = types.Int64Value(int64(createdSec.LatestVersion))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *secretManagerSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data secretManagerSecretResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, data.Timeouts, common.Timeout5min)
	defer cancel()

	secretOp := sm.NewSecretOp(r.client, data.VaultID.ValueString())

	_, err := FilterSecretManagerSecretByName(ctx, secretOp, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("SecretManagerSecret Delete's Read Error", err.Error())
		return
	}

	err = secretOp.Delete(ctx, v1.DeleteSecret{Name: data.Name.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("SecretManagerSecret Delete Error", err.Error())
		return
	}
}

func FilterSecretManagerSecretByName(ctx context.Context, secretOp sm.SecretAPI, name string) (*v1.Secret, error) {
	secrets, err := secretOp.List(ctx)
	if err != nil {
		return nil, err
	}

	match := slices.Collect(func(yield func(v1.Secret) bool) {
		for _, v := range secrets {
			if name != v.Name {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})

	if len(match) == 0 {
		return nil, common.ErrFilterNoResult
	}

	return &match[0], nil
}
