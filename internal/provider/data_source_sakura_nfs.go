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
	iaastypes "github.com/sacloud/iaas-api-go/types"
)

type nfsDataSource struct {
	client *APIClient
}

var (
	_ datasource.DataSource              = &nfsDataSource{}
	_ datasource.DataSourceWithConfigure = &nfsDataSource{}
)

func NewNFSDataSource() datasource.DataSource {
	return &nfsDataSource{}
}

func (d *nfsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nfs"
}

func (d *nfsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type nfsDataSourceModel struct {
	sakuraNFSBaseModel
}

func (d *nfsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaDataSourceId("NFS"),
			"name":        schemaDataSourceName("NFS"),
			"description": schemaDataSourceDescription("NFS"),
			"tags":        schemaDataSourceTags("NFS"),
			"zone":        schemaDataSourceZone("NFS"),
			"icon_id":     schemaDataSourceIconID("NFS"),
			"plan":        schemaDataSourcePlan("NFS", iaastypes.NFSPlanStrings),
			"size":        schemaDataSourceSize("NFS"),
			"network_interface": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"switch_id":  schemaDataSourceSwitchID("NFS"),
						"ip_address": schemaDataSourceIPAddress("NFS"),
						"netmask":    schemaDataSourceNetMask("NFS"),
						"gateway":    schemaDataSourceGateway("NFS"),
					},
				},
			},
		},
	}
}

func (d *nfsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data nfsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewNFSOp(d.client)
	findCondition := createFindCondition(data.ID, data.Name, data.Tags)

	res, err := searcher.Find(ctx, zone, findCondition)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud NFS resource: %s", err))
		return
	}
	if res == nil || res.Count == 0 || len(res.NFS) == 0 {
		filterNoResultErr(&resp.Diagnostics)
		return
	}

	if _, err := data.updateState(ctx, d.client, res.NFS[0], zone); err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not update state for SakuraCloud NFS resource: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
