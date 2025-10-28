// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package secret_manager

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	sm "github.com/sacloud/secretmanager-api-go"
	v1 "github.com/sacloud/secretmanager-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type secretManagerDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &secretManagerDataSource{}
	_ datasource.DataSourceWithConfigure = &secretManagerDataSource{}
)

func NewSecretManagerDataSource() datasource.DataSource {
	return &secretManagerDataSource{}
}

func (d *secretManagerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret_manager"
}

func (d *secretManagerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.SecretManagerClient
}

type secretManagerDataSourceModel struct {
	secretManagerBaseModel
}

func (d *secretManagerDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("SecretManager vault"),
			"description": common.SchemaDataSourceDescription("SecretManager vault"),
			"tags":        common.SchemaDataSourceTags("SecretManager vault"),
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the SecretManager vault.",
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
	if !data.Name.IsNull() { //nolint:gocritic
		vaults, err := vaultOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("SecretManager List Error", err.Error())
			return
		}
		vault, err = FilterSecretManagerVaultByName(vaults, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("SecretManager Filter Error", err.Error())
			return
		}
	} else if !data.ID.IsNull() {
		vault, err = vaultOp.Read(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("SecretManager Read Error", err.Error())
			return
		}
	} else {
		resp.Diagnostics.AddError("Missing Attribute", "Either 'id' or 'name' must be specified.")
		return
	}

	data.updateState(vault)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FilterSecretManagerVaultByName(vaults []v1.Vault, name string) (*v1.Vault, error) {
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
