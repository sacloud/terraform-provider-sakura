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

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/query"
	"github.com/sacloud/iaas-api-go/ostype"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

type archiveDataSource struct {
	client *APIClient
}

var (
	_ datasource.DataSource              = &archiveDataSource{}
	_ datasource.DataSourceWithConfigure = &archiveDataSource{}
)

func NewArchiveDataSource() datasource.DataSource {
	return &archiveDataSource{}
}

type archiveDataSourceModel struct {
	sakuraBaseModel
	Zone   types.String `tfsdk:"zone"`
	Size   types.Int64  `tfsdk:"size"`
	OSType types.String `tfsdk:"os_type"`
	IconID types.String `tfsdk:"icon_id"`
}

func (d *archiveDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_archive"
}

func (d *archiveDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

func (d *archiveDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaDataSourceId("Archive"),
			"name":        schemaDataSourceName("Archive"),
			"description": schemaDataSourceDescription("Archive"),
			"tags":        schemaDataSourceTags("Archive"),
			"zone":        schemaDataSourceZone("Archive"),
			"icon_id":     schemaDataSourceIconID("Archive"),
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "The size of the archive in GB.",
			},
			"os_type": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The criteria used to filter SakuraCloud archives. This must be one of following: \n%s", ostype.OSTypeShortNames),
				Validators: []validator.String{
					stringvalidator.OneOf(ostype.OSTypeShortNames...),
				},
			},
			// filter機能はv3から削除
		},
	}
}

func (d *archiveDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data archiveDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	zone := getZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewArchiveOp(d.client)
	strOSType := data.OSType.ValueString()
	archive, err := query.FindArchiveByOSType(ctx, searcher, zone, ostype.StrToOSType(strOSType))
	if err != nil {
		resp.Diagnostics.AddError("Archive Search Error", err.Error())
		return
	}

	data.updateBaseState(archive.ID.String(), archive.Name, archive.Description, archive.Tags)
	data.Size = types.Int64Value(int64(archive.GetSizeGB()))
	data.IconID = types.StringValue(archive.IconID.String())
	data.Zone = types.StringValue(zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
