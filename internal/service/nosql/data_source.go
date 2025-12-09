// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package nosql

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/nosql-api-go"
	v1 "github.com/sacloud/nosql-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type nosqlDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &nosqlDataSource{}
	_ datasource.DataSourceWithConfigure = &nosqlDataSource{}
)

func NewNosqlDataSource() datasource.DataSource {
	return &nosqlDataSource{}
}

func (d *nosqlDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nosql"
}

func (d *nosqlDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.NosqlClient
}

type nosqlDataSourceModel struct {
	nosqlBaseModel
}

func (d *nosqlDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("NoSQL appliance"),
			"name":        common.SchemaDataSourceName("NoSQL appliance"),
			"description": common.SchemaDataSourceDescription("NoSQL appliance"),
			"tags":        common.SchemaDataSourceTags("NoSQL appliance"),
			"zone":        common.SchemaDataSourceZone("NoSQL appliance"),
			"plan":        common.SchemaDataSourcePlan("NoSQL appliance", nosql.Plan100GB.AllValuesAsString()),
			"settings": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Settings of the NoSQL appliance",
				Attributes: map[string]schema.Attribute{
					"source_network": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "Source network address",
					},
					"reserve_ip_address": schema.StringAttribute{
						Computed:    true,
						Description: "Reserved IP address. This address is used for dead node replacement",
					},
					"backup": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"connect": schema.StringAttribute{
								Computed:    true,
								Description: "Backup destination by NFS URL format",
							},
							"days_of_week": schema.SetAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "Backup schedule",
							},
							"time": schema.StringAttribute{
								Computed:    true,
								Description: "Time for backup execution",
							},
							"rotate": schema.Int32Attribute{
								Computed:    true,
								Description: "Number of backup rotations",
							},
						},
					},
					"repair": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Regular repair configuration",
						Attributes: map[string]schema.Attribute{
							"incremental": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "Incremental repair configuration",
								Attributes: map[string]schema.Attribute{
									"days_of_week": schema.SetAttribute{
										ElementType: types.StringType,
										Computed:    true,
										Description: "Repair schedule",
									},
									"time": schema.StringAttribute{
										Computed:    true,
										Description: "Time for incremental repair execution",
									},
								},
							},
							"full": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "Full repair configuration",
								Attributes: map[string]schema.Attribute{
									"interval": schema.Int32Attribute{
										Computed:    true,
										Description: "Execution interval of 7 days. e.g. 7 / 14 / 21 / 28",
									},
									"day_of_week": schema.StringAttribute{
										Computed:    true,
										Description: "Repair schedule",
									},
									"time": schema.StringAttribute{
										Computed:    true,
										Description: "Time for full repair execution",
									},
								},
							},
						},
					},
				},
			},
			"remark": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"nosql": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "NoSQL database information",
						Attributes: map[string]schema.Attribute{
							"engine": schema.StringAttribute{
								Computed:    true,
								Description: "Database engine used by NoSQL appliance.",
							},
							"version": schema.StringAttribute{
								Computed:    true,
								Description: "Version of database engine used by NoSQL appliance.",
							},
							"default_user": schema.StringAttribute{
								Computed:    true,
								Description: "Default user for NoSQL appliance.",
							},
							"disk_size": schema.Int32Attribute{
								Computed:    true,
								Description: "Disk size of NoSQL appliance.",
							},
							"memory": schema.Int32Attribute{
								Computed:    true,
								Description: "Memory size of NoSQL appliance.",
							},
							"nodes": schema.Int32Attribute{
								Computed:    true,
								Description: "Number of nodes. 3 for primary node, 2 for additional nodes",
							},
							"port": schema.Int32Attribute{
								Computed:    true,
								Description: "Port number used by NoSQL appliance.",
							},
							"virtualcore": schema.Int32Attribute{
								Computed:    true,
								Description: "Number of virtual cores used by NoSQL appliance.",
							},
							"zone": schema.StringAttribute{
								Computed:    true,
								Description: "Zone where NoSQL appliance is located.",
							},
							"primary_nodes": schema.SingleNestedAttribute{
								Computed:    true,
								Description: "Primary Node information. This field is used in additional nodes",
								Attributes: map[string]schema.Attribute{
									"id":   schema.StringAttribute{Computed: true},
									"zone": schema.StringAttribute{Computed: true},
								},
							},
						},
					},
					"servers": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "IP addresses which connect to user's switch",
					},
					"network": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"gateway": schema.StringAttribute{
								Computed:    true,
								Description: "Gateway address of the network",
							},
							"netmask": schema.Int32Attribute{
								Computed:    true,
								Description: "Netmask of the network",
							},
						},
					},
					"zone_id": schema.StringAttribute{
						Computed:    true,
						Description: "The id of the zone where NoSQL appliance is located.",
					},
				},
			},
			"instance": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Instance and host information",
				Attributes: map[string]schema.Attribute{
					"status": schema.StringAttribute{
						Computed:    true,
						Description: "The NoSQL instance status. 'up' or 'down'",
					},
					"status_changed_at": schema.StringAttribute{
						Computed:    true,
						Description: "The time when the status was last changed",
					},
					"host": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed:    true,
								Description: "The host name where NoSQL appliance is running",
							},
							"info_url": schema.StringAttribute{
								Computed:    true,
								Description: "The information URL of the host where NoSQL appliance is running",
							},
						},
					},
					"hosts": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "The host name where NoSQL appliance is running",
								},
								"info_url": schema.StringAttribute{
									Computed:    true,
									Description: "The information URL of the host where NoSQL appliance is running",
								},
							},
						},
					},
				},
			},
			"disk": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Disk encryption information",
				Attributes: map[string]schema.Attribute{
					"encryption_key": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Encryption key setting. Specify KMS key ID.",
						Attributes: map[string]schema.Attribute{
							"kms_key_id": schema.StringAttribute{
								Computed:    true,
								Description: "KMS key ID for disk encryption",
							},
						},
					},
					"encryption_algorithm": schema.StringAttribute{
						Computed:    true,
						Description: "Encryption algorithm used for disk encryption",
					},
				},
			},
			"interfaces": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Network interfaces",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip_address": schema.StringAttribute{
							Computed:    true,
							Description: "IP Address assigned to the interface",
						},
						"user_ip_address": schema.StringAttribute{
							Computed:    true,
							Description: "IP Address which connect to user's switch",
						},
						"hostname": schema.StringAttribute{
							Computed:    true,
							Description: "Hostname assigned to the interface",
						},
						"vswitch": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Computed:    true,
									Description: "The ID of the vSwitch connected to the interface",
								},
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "The name of the vSwitch connected to the interface",
								},
								"scope": schema.StringAttribute{
									Computed:    true,
									Description: "The scope of the vSwitch connected to the interface",
								},
								"subnet": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"network_address": schema.StringAttribute{
											Computed:    true,
											Description: "The network address of the subnet connected to the interface",
										},
										"netmask": schema.Int32Attribute{
											Computed:    true,
											Description: "The netmask of the subnet connected to the interface",
										},
										"gateway": schema.StringAttribute{
											Computed:    true,
											Description: "The gateway of the subnet connected to the interface",
										},
										"band_width_mbps": schema.Int32Attribute{
											Computed:    true,
											Description: "The bandwidth in Mbps of the subnet connected to the interface",
										},
									},
								},
								"user_subnet": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"netmask": schema.Int32Attribute{
											Computed:    true,
											Description: "The netmask of the user subnet connected to the interface",
										},
										"gateway": schema.StringAttribute{
											Computed:    true,
											Description: "The gateway of the user subnet connected to the interface",
										},
									},
								},
							},
						},
					},
				},
			},
			"availability": schema.StringAttribute{
				Computed:    true,
				Description: "Availability state. state is one of migrating / available / failed",
			},
			"generation": schema.Int32Attribute{
				Computed:    true,
				Description: "Generation number",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation time",
			},
		},
		MarkdownDescription: "Get information about an existing NoSQL Database.",
	}
}

func (d *nosqlDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data nosqlDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseOp := nosql.NewDatabaseOp(d.client)
	var res *v1.GetNosqlAppliance
	var err error
	if !data.Name.IsNull() {
		nosqls, err := databaseOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("NoSQL List Error", fmt.Sprintf("could not find NoSQL resource: %s", err))
			return
		}
		res, err = filterNosqlByName(nosqls, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("NoSQL Filter Error", err.Error())
			return
		}
	} else {
		res, err = databaseOp.Read(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("NoSQL Read Error", "No result found")
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterNosqlByName(nosqls []v1.GetNosqlAppliance, name string) (*v1.GetNosqlAppliance, error) {
	match := slices.Collect(func(yield func(v1.GetNosqlAppliance) bool) {
		for _, v := range nosqls {
			if name != v.Name.Value {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple NoSQL resources found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
