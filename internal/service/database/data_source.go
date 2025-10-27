// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package database

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

type databaseDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &databaseDataSource{}
	_ datasource.DataSourceWithConfigure = &databaseDataSource{}
)

func NewDatabaseDataSource() datasource.DataSource {
	return &databaseDataSource{}
}

func (d *databaseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (d *databaseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type databaseDataSourceModel struct {
	databaseBaseModel
}

func (d *databaseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Database"),
			"name":        common.SchemaDataSourceName("Database"),
			"description": common.SchemaDataSourceDescription("Database"),
			"tags":        common.SchemaDataSourceTags("Database"),
			"icon_id":     common.SchemaDataSourceIconID("Database"),
			"zone":        common.SchemaDataSourceZone("Database"),
			"plan":        common.SchemaDataSourcePlan("Database", iaastypes.DatabasePlanStrings),
			"database_type": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The type of the database. This will be one of [%s]", iaastypes.RDBMSTypeStrings),
			},
			"database_version": schema.StringAttribute{
				Computed:    true,
				Description: "The version of the database",
			},
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "The name of default user on the database",
			},
			"password": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The password of default user on the database",
			},
			"replica_user": schema.StringAttribute{
				Computed:    true,
				Description: "The name of user that processing a replication",
			},
			"replica_password": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The password of user that processing a replication",
			},
			"network_interface": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Network interfaces (simplified map form)",
				Attributes: map[string]schema.Attribute{
					"switch_id":  common.SchemaDataSourceSwitchID("Database"),
					"ip_address": common.SchemaDataSourceIPAddress("Database"),
					"netmask":    common.SchemaDataSourceNetMask("Database"),
					"gateway":    common.SchemaDataSourceGateway("Database"),
					"port": schema.Int32Attribute{
						Computed:    true,
						Description: "The number of the listening port",
					},
					"source_ranges": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "The range of source IP addresses that allow to access to the Database via network",
					},
				},
			},
			"backup": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Backup settings (simplified)",
				Attributes: map[string]schema.Attribute{
					"weekdays": schema.SetAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: desc.Sprintf("The list of name of weekday that doing backup. This will be in [%s]", iaastypes.DaysOfTheWeekStrings),
					},
					"time": schema.StringAttribute{
						Computed:    true,
						Description: "The time to take backup. This will be formatted with `HH:mm`",
					},
				},
			},
			"parameters": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The map for setting RDBMS-specific parameters. Valid keys can be found with the `usacloud database list-parameters` command",
			},
		},
	}
}

func (d *databaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data databaseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewDatabaseOp(d.client)
	res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to find SakuraCloud Database: %s", err))
		return
	}
	if res == nil || res.Count == 0 || len(res.Databases) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	db := res.Databases[0]
	if removeDB, err := data.updateState(ctx, d.client, zone, db); err != nil {
		if removeDB {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to update SakuraCloud Database[%s] state: %s", db.ID.String(), err))
		return
	}
	data.IconID = types.StringValue(db.IconID.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
