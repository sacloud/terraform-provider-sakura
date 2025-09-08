// Copyright 2016-2025 terraform-provider-sakura authors
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

package icon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type iconDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &iconDataSource{}
	_ datasource.DataSourceWithConfigure = &iconDataSource{}
)

func NewIconDataSource() datasource.DataSource {
	return &iconDataSource{}
}

func (d *iconDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_icon"
}

func (d *iconDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type iconDataSourceModel struct {
	iconBaseModel
}

func (d *iconDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaDataSourceId("Icon"),
			"name": common.SchemaDataSourceName("Icon"),
			"tags": common.SchemaDataSourceTags("Icon"),
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL for getting the icon's raw data",
			},
		},
	}
}

func (d *iconDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data iconDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewIconOp(d.client)
	res, err := searcher.Find(ctx, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", "could not find SakuraCloud Icon")
		return
	}
	if res == nil || res.Count == 0 || len(res.Icons) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.Icons[0])
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
