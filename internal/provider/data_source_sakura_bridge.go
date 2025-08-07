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

type bridgeDataSourceModel struct {
	bridgeResourceModel
	Filter *filterBlockModel `tfsdk:"filter"`
}

type bridgeDataSource struct {
	client *APIClient
}

var _ datasource.DataSource = &bridgeDataSource{}

func NewBridgeDataSource() datasource.DataSource {
	return &bridgeDataSource{}
}

func (d *bridgeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bridge"
}

func (d *bridgeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

func (d *bridgeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaDataSourceId("Bridge"),
			"name":        schemaDataSourceName("Bridge"),
			"description": schemaDataSourceDescription("Bridge"),
			"zone":        schemaDataSourceZone("Bridge"),
		},
		Blocks: filterSchema(&filterSchemaOption{}),
	}
}

func (d *bridgeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data bridgeDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	bridgeOp := iaas.NewBridgeOp(d.client)
	findCondition := &iaas.FindCondition{}
	if data.Filter != nil {
		findCondition.Filter = expandSearchFilter(data.Filter)
	}
	res, err := bridgeOp.Find(ctx, zone, findCondition)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud Bridge : %s", err))
		return
	}
	if res == nil || len(res.Bridges) == 0 {
		filterNoResultErr(&resp.Diagnostics)
		return
	}

	bridge := res.Bridges[0]
	data.updateState(ctx, bridge, zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
