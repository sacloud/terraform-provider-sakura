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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/iaas-api-go"
)

type sshKeyDataSource struct {
	client *APIClient
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
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type sshKeyDataSourceModel struct {
	sakuraSSHKeyBaseModel
	Filter *filterBlockModel `tfsdk:"filter"`
}

func (d *sshKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaDataSourceId("SSHKey"),
			"name":        schemaDataSourceName("SSHKey"),
			"description": schemaDataSourceDescription("SSHKey"),
			"public_key": schema.StringAttribute{
				Computed:    true,
				Description: "The value of public key",
			},
			"fingerprint": schema.StringAttribute{
				Computed:    true,
				Description: "The fingerprint of public key",
			},
		},
		Blocks: filterSchema(&filterSchemaOption{}),
	}
}

func (d *sshKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sshKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewSSHKeyOp(d.client)
	findCondition := &iaas.FindCondition{}
	if data.Filter != nil {
		findCondition.Filter = expandSearchFilter(data.Filter)
	}

	res, err := searcher.Find(ctx, findCondition)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud SSHKey resource: %s", err.Error()))
		return
	}
	if res == nil || res.Count == 0 || len(res.SSHKeys) == 0 {
		filterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.SSHKeys[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
