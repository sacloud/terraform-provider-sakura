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
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

type diskDataSource struct {
	client *APIClient
}

var (
	_ datasource.DataSource              = &diskDataSource{}
	_ datasource.DataSourceWithConfigure = &diskDataSource{}
)

func NewDiskDataSource() datasource.DataSource {
	return &diskDataSource{}
}

func (d *diskDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_disk"
}

func (d *diskDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type diskDataSourceModel struct {
	sakuraDiskBaseModel
}

func (d *diskDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaDataSourceId("Disk"),
			"name":        schemaDataSourceName("Disk"),
			"description": schemaDataSourceDescription("Disk"),
			"tags":        schemaDataSourceTags("Disk"),
			"icon_id":     schemaDataSourceIconID("Disk"),
			"zone":        schemaDataSourceZone("Disk"),
			"size":        schemaDataSourceSize("Disk"),
			"plan":        schemaDataSourcePlan("Disk", iaastypes.DiskPlanStrings),
			"server_id":   schemaDataSourceServerID("Disk"),
			"connector": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The name of the disk connector. This will be one of [%s]", iaastypes.DiskConnectionStrings),
			},
			"encryption_algorithm": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The disk encryption algorithm. This must be one of [%s]", iaastypes.DiskEncryptionAlgorithmStrings),
			},
			"source_archive_id": schema.StringAttribute{
				Computed:    true,
				Description: "The id of the source archive",
			},
			"source_disk_id": schema.StringAttribute{
				Computed:    true,
				Description: "The id of the source disk",
			},
		},
	}
}

func (d *diskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data diskDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone := getZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewDiskOp(d.client)
	res, err := searcher.Find(ctx, zone, createFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud Disk resource: %s", err))
		return
	}
	if res == nil || res.Count == 0 || len(res.Disks) == 0 {
		filterNoResultErr(&resp.Diagnostics)
		return
	}

	data.updateState(res.Disks[0], zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
