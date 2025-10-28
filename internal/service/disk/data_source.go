// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package disk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type diskDataSource struct {
	client *common.APIClient
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
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type diskDataSourceModel struct {
	diskBaseModel
}

func (d *diskDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Disk"),
			"name":        common.SchemaDataSourceName("Disk"),
			"description": common.SchemaDataSourceDescription("Disk"),
			"tags":        common.SchemaDataSourceTags("Disk"),
			"icon_id":     common.SchemaDataSourceIconID("Disk"),
			"zone":        common.SchemaDataSourceZone("Disk"),
			"size":        common.SchemaDataSourceSize("Disk"),
			"plan":        common.SchemaDataSourcePlan("Disk", iaastypes.DiskPlanStrings),
			"server_id":   common.SchemaDataSourceServerID("Disk"),
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
			"kms_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the KMS key for encryption",
			},
		},
		MarkdownDescription: "Get information about an existing Disk.",
	}
}

func (d *diskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data diskDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewDiskOp(d.client)
	res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud Disk resource: %s", err))
		return
	}
	if res == nil || res.Count == 0 || len(res.Disks) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	disk := res.Disks[0]
	data.updateState(disk, zone)
	data.IconID = types.StringValue(disk.IconID.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
