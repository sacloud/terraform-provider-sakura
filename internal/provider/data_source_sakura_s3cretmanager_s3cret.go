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

package sakura

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
)

type secretManagerSecretDataSourceModel struct {
	Name    types.String `tfsdk:"name"`
	VaultID types.String `tfsdk:"vault_id"`
	Version types.Int64  `tfsdk:"version"`
	Value   types.String `tfsdk:"value"`
}

func NewSecretManagerSecretDataSource() datasource.DataSource {
	return &secretManagerSecretDataSource{}
}

type secretManagerSecretDataSource struct {
	client *v1.Client
}

func (d *secretManagerSecretDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

var _ datasource.DataSource = &secretManagerSecretDataSource{}
var _ datasource.DataSourceWithConfigure = &secretManagerSecretDataSource{}

func (d *secretManagerSecretDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secretmanager_secret"
}

func (d *secretManagerSecretDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the secret.",
			},
			"vault_id": schema.StringAttribute{
				Required:    true,
				Description: "The secret manager's vault id.",
			},
			"version": schema.Int64Attribute{
				Optional:    true,
				Description: "Target version to unveil stored secret. Without this parameter, latest version is used.",
			},
			"value": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Unveiled result of stored secret.",
			},
		},
	}
}

func (d *secretManagerSecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data secretManagerSecretDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secretOp := sm.NewSecretOp(d.client, data.VaultID.ValueString())

	unveilReq := v1.Unveil{Name: data.Name.ValueString()}
	if !data.Version.IsNull() {
		unveilReq.Version = v1.NewOptNilInt(int(data.Version.ValueInt64()))
	}
	unveil, err := secretOp.Unveil(ctx, unveilReq)
	if err != nil {
		resp.Diagnostics.AddError("SecretManagerSecret Unveil Error", err.Error())
		return
	}

	data.Name = types.StringValue(unveil.Name)
	data.Value = types.StringValue(unveil.Value)
	if unveil.Version.IsSet() {
		data.Version = types.Int64Value(int64(unveil.Version.Value))
	}

	resp.State.Set(ctx, &data)
}
