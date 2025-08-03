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
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

type secretManagerDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ResourceID  types.String `tfsdk:"resource_id"`
	Description types.String `tfsdk:"description"`
	Tags        types.Set    `tfsdk:"tags"`
	KmsKeyID    types.String `tfsdk:"kms_key_id"`
}

func NewSecretManagerDataSource() datasource.DataSource {
	return &secretManagerDataSource{}
}

type secretManagerDataSource struct {
	client *v1.Client
}

func (d *secretManagerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	apiclient, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected ProviderData type", "Expected *APIClient.")
		return
	}
	d.client = apiclient.secretmanagerClient
}

var _ datasource.DataSource = &secretManagerDataSource{}
var _ datasource.DataSourceWithConfigure = &secretManagerDataSource{}

func (d *secretManagerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_manager"
}

func (d *secretManagerDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaDataSourceId("SecretManager vault"),
			"description": schemaDataSourceDescription("SecretManager vault"),
			"tags":        schemaDataSourceTags("SecretManager vault"),
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the SecretManager vault.",
			},
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Description: "The resource ID of the SecretManager vault.",
			},
			"kms_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "KMS key id for the SecretManager vault.",
			},
		},
	}
}

func (d *secretManagerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data secretManagerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vaultOp := sm.NewVaultOp(d.client)

	var vault *v1.Vault
	var err error
	if !data.Name.IsNull() {
		vaults, err := vaultOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("SecretManager List Error", err.Error())
			return
		}
		vault, err = filterSecretManagerVaultByName(vaults, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("SecretManager Filter Error", err.Error())
			return
		}
	} else if !data.ResourceID.IsNull() {
		vault, err = vaultOp.Read(ctx, data.ResourceID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("SecretManager Read Error", err.Error())
			return
		}
	} else {
		resp.Diagnostics.AddError("Missing Attribute", "Either 'name' or 'resource_id' must be specified.")
		return
	}

	data.ID = types.StringValue(vault.ID)
	data.Name = types.StringValue(vault.Name)
	data.Description = types.StringValue(vault.Description.Value)
	data.Tags = TagsToTFSet(ctx, vault.Tags)
	data.KmsKeyID = types.StringValue(vault.KmsKeyID)

	resp.State.Set(ctx, &data)
}

func filterSecretManagerVaultByName(vaults []v1.Vault, name string) (*v1.Vault, error) {
	match := slices.Collect(func(yield func(v1.Vault) bool) {
		for _, v := range vaults {
			if name != v.Name {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple SecretManager vault resources found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
