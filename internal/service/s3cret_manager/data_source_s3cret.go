// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package secret_manager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type secretManagerSecretDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &secretManagerSecretDataSource{}
	_ datasource.DataSourceWithConfigure = &secretManagerSecretDataSource{}
)

func NewSecretManagerSecretDataSource() datasource.DataSource {
	return &secretManagerSecretDataSource{}
}

func (d *secretManagerSecretDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_manager_secret"
}

func (d *secretManagerSecretDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.SecretManagerClient
}

type secretManagerSecretDataSourceModel struct {
	secretManagerSecretBaseModel
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
		MarkdownDescription: "Get information about an existing Secret Manager's secret.",
	}
}

func (d *secretManagerSecretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data secretManagerSecretDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	unveilReq := v1.Unveil{Name: data.Name.ValueString()}
	if !data.Version.IsNull() {
		unveilReq.Version = v1.NewOptNilInt(int(data.Version.ValueInt64()))
	}

	secretOp := sm.NewSecretOp(d.client, data.VaultID.ValueString())
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
