// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package archive

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/query"
	"github.com/sacloud/iaas-api-go/ostype"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type archiveDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &archiveDataSource{}
	_ datasource.DataSourceWithConfigure = &archiveDataSource{}
)

func NewArchiveDataSource() datasource.DataSource {
	return &archiveDataSource{}
}

func (d *archiveDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_archive"
}

func (d *archiveDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type archiveDataSourceModel struct {
	common.SakuraBaseModel
	Zone   types.String `tfsdk:"zone"`
	Size   types.Int64  `tfsdk:"size"`
	OSType types.String `tfsdk:"os_type"`
	IconID types.String `tfsdk:"icon_id"`
}

func (d *archiveDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Archive"),
			"name":        common.SchemaDataSourceName("Archive"),
			"description": common.SchemaDataSourceDescription("Archive"),
			"tags":        common.SchemaDataSourceTags("Archive"),
			"zone":        common.SchemaDataSourceZone("Archive"),
			"icon_id":     common.SchemaDataSourceIconID("Archive"),
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "The size of the archive in GB.",
			},
			"os_type": schema.StringAttribute{
				Optional:    true,
				Description: desc.Sprintf("The criteria used to filter SakuraCloud archives. This must be one of following: \n%s", ostype.OSTypeShortNames),
				Validators: []validator.String{
					stringvalidator.OneOf(ostype.OSTypeShortNames...),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("id"),
						path.MatchRelative().AtParent().AtName("name"), path.MatchRelative().AtParent().AtName("tags")),
				},
			},
		},
		MarkdownDescription: "Get information about an existing Archive.",
	}
}

func (d *archiveDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data archiveDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewArchiveOp(d.client)
	var archive *iaas.Archive
	if !data.OSType.IsNull() && !data.OSType.IsUnknown() {
		strOSType := data.OSType.ValueString()
		res, err := query.FindArchiveByOSType(ctx, searcher, zone, ostype.StrToOSType(strOSType))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find Archive by OS type: %s", err))
			return
		}
		archive = res
	} else {
		res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find Archive: %s", err))
			return
		}
		if res == nil || len(res.Archives) == 0 {
			common.FilterNoResultErr(&resp.Diagnostics)
			return
		}
		archive = res.Archives[0]
	}

	data.UpdateBaseState(archive.ID.String(), archive.Name, archive.Description, archive.Tags)
	data.Size = types.Int64Value(int64(archive.GetSizeGB()))
	data.IconID = types.StringValue(archive.IconID.String())
	data.Zone = types.StringValue(zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
