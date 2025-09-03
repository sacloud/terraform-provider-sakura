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

package sw1tch

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
)

type switchDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &switchDataSource{}
	_ datasource.DataSourceWithConfigure = &switchDataSource{}
)

func NewSwitchDataSource() datasource.DataSource {
	return &switchDataSource{}
}

func (d *switchDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch"
}

func (d *switchDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type switchDataSourceModel struct {
	switchBaseModel
}

func (d *switchDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Switch"),
			"name":        common.SchemaDataSourceName("Switch"),
			"description": common.SchemaDataSourceDescription("Switch"),
			"tags":        common.SchemaDataSourceTags("Switch"),
			"icon_id":     common.SchemaDataSourceIconID("Switch"),
			"zone":        common.SchemaDataSourceZone("Switch"),
			"bridge_id": schema.StringAttribute{
				Computed:    true,
				Description: "The bridge id attached to the Switch.",
			},
			"server_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A set of server id connected to the Switch",
			},
		},
	}
}

func (d *switchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data switchDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewSwitchOp(d.client)
	res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.Switches) == 0 {
		resp.Diagnostics.AddError("Read Error", "No SakuraCloud Switch resource matched the filter.")
		return
	}

	sw := res.Switches[0]
	if err := data.updateState(ctx, d.client, sw, zone); err != nil {
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}
	data.IconID = types.StringValue(sw.IconID.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
