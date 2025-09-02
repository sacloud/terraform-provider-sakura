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

package private_host

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
)

// PrivateHostDataSource implements datasource.DataSource
type privateHostDataSource struct {
	client *common.APIClient
}

// Ensure privateHostDataSource implements the required interfaces
var (
	_ datasource.DataSource              = &privateHostDataSource{}
	_ datasource.DataSourceWithConfigure = &privateHostDataSource{}
)

// NewPrivateHostDataSource returns a new instance
func NewPrivateHostDataSource() datasource.DataSource {
	return &privateHostDataSource{}
}

func (d *privateHostDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_host"
}

func (d *privateHostDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type privateHostDataSourceModel struct {
	privateHostBaseModel
}

func (d *privateHostDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("PrivateHost"),
			"name":        common.SchemaDataSourceName("PrivateHost"),
			"description": common.SchemaDataSourceDescription("PrivateHost"),
			"tags":        common.SchemaDataSourceTags("PrivateHost"),
			"zone":        common.SchemaDataSourceZone("PrivateHost"),
			"icon_id":     common.SchemaDataSourceIconID("PrivateHost"),
			"class":       common.SchemaDataSourceClass("PrivateHost", []string{iaastypes.PrivateHostClassDynamic, iaastypes.PrivateHostClassWindows}),
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "The hostname of the private host.",
			},
			"assigned_core": schema.Int32Attribute{
				Computed:    true,
				Description: "The total number of CPUs assigned to servers on the private host",
			},
			"assigned_memory": schema.Int32Attribute{
				Computed:    true,
				Description: "The total size of memory assigned to servers on the private host",
			},
		},
	}
}

func (d *privateHostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data privateHostDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewPrivateHostOp(d.client)
	findCondition := common.CreateFindCondition(data.ID, data.Name, data.Tags)

	res, err := searcher.Find(ctx, zone, findCondition)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.PrivateHosts) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	ph := res.PrivateHosts[0]
	data.updateState(ph, zone)
	data.IconID = types.StringValue(ph.IconID.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// sakuraCloudClientFramework, expandSearchFilterFrameworkはSDK/Framework用に実装してください
