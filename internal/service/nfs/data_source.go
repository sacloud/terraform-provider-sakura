// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package nfs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type nfsDataSource struct {
	client *common.APIClient
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
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type nfsDataSourceModel struct {
	nfsBaseModel
}

func (d *nfsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("NFS"),
			"name":        common.SchemaDataSourceName("NFS"),
			"description": common.SchemaDataSourceDescription("NFS"),
			"tags":        common.SchemaDataSourceTags("NFS"),
			"zone":        common.SchemaDataSourceZone("NFS"),
			"icon_id":     common.SchemaDataSourceIconID("NFS"),
			"plan":        common.SchemaDataSourcePlan("NFS", iaastypes.NFSPlanStrings),
			"size":        common.SchemaDataSourceSize("NFS"),
			"network_interface": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"vswitch_id": common.SchemaDataSourceVSwitchID("NFS"),
					"ip_address": common.SchemaDataSourceIPAddress("NFS"),
					"netmask":    common.SchemaDataSourceNetMask("NFS"),
					"gateway":    common.SchemaDataSourceGateway("NFS"),
				},
			},
		},
		MarkdownDescription: "Get information about an existing NFS.",
	}
}

func (d *nfsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data nfsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewNFSOp(d.client)
	findCondition := common.CreateFindCondition(data.ID, data.Name, data.Tags)

	res, err := searcher.Find(ctx, zone, findCondition)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud NFS resource: %s", err))
		return
	}
	if res == nil || res.Count == 0 || len(res.NFS) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	nfs := res.NFS[0]
	if _, err := data.updateState(ctx, d.client, nfs, zone); err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not update state for SakuraCloud NFS resource: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
