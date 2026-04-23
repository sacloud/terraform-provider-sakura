// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ondemand_db

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	enhanceddbbuilder "github.com/sacloud/iaas-service-go/enhanceddb/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type onDemandDBDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &onDemandDBDataSource{}
	_ datasource.DataSourceWithConfigure = &onDemandDBDataSource{}
)

func NewOnDemandDBDataSource() datasource.DataSource {
	return &onDemandDBDataSource{}
}

func (d *onDemandDBDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ondemand_db"
}

func (d *onDemandDBDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiClient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiClient == nil {
		return
	}
	d.client = apiClient
}

type onDemandDBDataSourceModel struct {
	onDemandDBBaseModel
}

func (d *onDemandDBDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resourceName := "OnDemand Database"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId(resourceName),
			"name":        common.SchemaDataSourceName(resourceName),
			"description": common.SchemaDataSourceDescription(resourceName),
			"tags":        common.SchemaDataSourceTags(resourceName),
			"icon_id":     common.SchemaDataSourceIconID(resourceName),
			"database_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of database",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "The region name",
			},
			"database_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of database",
			},
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "The name of database host. This will be built from `database_name` + `tidb-is1.db.sakurausercontent.com`",
			},
			"allowed_networks": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: desc.Sprintf("A list of CIDR blocks allowed to connect"),
			},
			"max_connections": schema.Int64Attribute{
				Computed:    true,
				Description: "The value of max connections setting",
			},
		},
		MarkdownDescription: "Get information about an existing OnDemand Database.",
	}
}

func (d *onDemandDBDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data onDemandDBDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	edbOp := iaas.NewEnhancedDBOp(d.client)
	res, err := edbOp.Find(ctx, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", "failed to find OnDemand Database: "+err.Error())
		return
	}
	if res == nil || res.Count == 0 || len(res.EnhancedDBs) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	edb, err := enhanceddbbuilder.Read(ctx, edbOp, res.EnhancedDBs[0].ID)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", "failed to read OnDemand Database: "+err.Error())
		return
	}

	data.updateState(edb)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
