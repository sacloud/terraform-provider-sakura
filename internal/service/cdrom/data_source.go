// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cdrom

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type cdromDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &cdromDataSource{}
	_ datasource.DataSourceWithConfigure = &cdromDataSource{}
)

func NewCDROMDataSource() datasource.DataSource {
	return &cdromDataSource{}
}

func (d *cdromDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdrom"
}

func (d *cdromDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type cdromDataSourceModel struct {
	common.SakuraBaseModel
	Zone   types.String `tfsdk:"zone"`
	Size   types.Int64  `tfsdk:"size"`
	IconID types.String `tfsdk:"icon_id"`
}

func (d *cdromDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("CD-ROM"),
			"name":        common.SchemaDataSourceName("CD-ROM"),
			"description": common.SchemaDataSourceDescription("CD-ROM"),
			"tags":        common.SchemaDataSourceTags("CD-ROM"),
			"zone":        common.SchemaDataSourceZone("CD-ROM"),
			"icon_id":     common.SchemaDataSourceIconID("CD-ROM"),
			"size": schema.Int32Attribute{
				Computed:    true,
				Description: "The size of the CD-ROM in GiB.",
			},
		},
		MarkdownDescription: "Get information about an existing CD-ROM.",
	}
}

func (d *cdromDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cdromDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewCDROMOp(d.client)
	res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find CD-ROM: %s", err))
		return
	}
	if res == nil || len(res.CDROMs) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}
	cdrom := res.CDROMs[0]

	data.UpdateBaseState(cdrom.ID.String(), cdrom.Name, cdrom.Description, cdrom.Tags)
	data.Size = types.Int64Value(int64(cdrom.GetSizeGB()))
	data.IconID = types.StringValue(cdrom.IconID.String())
	data.Zone = types.StringValue(zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
