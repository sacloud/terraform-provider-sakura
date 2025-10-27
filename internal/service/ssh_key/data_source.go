// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ssh_key

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type sshKeyDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &sshKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &sshKeyDataSource{}
)

func NewSSHKeyDataSource() datasource.DataSource {
	return &sshKeyDataSource{}
}

func (d *sshKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (d *sshKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type sshKeyDataSourceModel struct {
	sshKeyBaseModel
}

func (d *sshKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("SSHKey"),
			"name":        common.SchemaDataSourceName("SSHKey"),
			"description": common.SchemaDataSourceDescription("SSHKey"),
			"public_key": schema.StringAttribute{
				Computed:    true,
				Description: "The value of public key",
			},
			"fingerprint": schema.StringAttribute{
				Computed:    true,
				Description: "The fingerprint of public key",
			},
		},
	}
}

func (d *sshKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sshKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewSSHKeyOp(d.client)
	res, err := searcher.Find(ctx, common.CreateFindCondition(data.ID, data.Name, types.SetNull(types.StringType)))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud SSHKey resource: %s", err.Error()))
		return
	}
	if res == nil || res.Count == 0 || len(res.SSHKeys) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.SSHKeys[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
