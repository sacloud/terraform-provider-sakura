// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package dsr_lb

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type dsrLBDataSource struct {
	client *common.APIClient
}

func NewDSRLBDataSource() datasource.DataSource {
	return &dsrLBDataSource{}
}

var (
	_ datasource.DataSource              = &dsrLBDataSource{}
	_ datasource.DataSourceWithConfigure = &dsrLBDataSource{}
)

func (r *dsrLBDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dsr_lb"
}

func (r *dsrLBDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type dsrLBDataSourceModel struct {
	dsrLBBaseModel
}

func (r *dsrLBDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("DSR LB"),
			"name":        common.SchemaDataSourceName("DSR LB"),
			"description": common.SchemaDataSourceDescription("DSR LB"),
			"tags":        common.SchemaDataSourceTags("DSR LB"),
			"zone":        common.SchemaDataSourceZone("DSR LB"),
			"icon_id":     common.SchemaDataSourceIconID("DSR LB"),
			"plan":        common.SchemaDataSourcePlan("DSR LB", []string{"standard", "highspec"}),
			"network_interface": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Network interface for DSR LB",
				Attributes: map[string]schema.Attribute{
					"vswitch_id":   common.SchemaDataSourceVSwitchID("DSR LB"),
					"ip_addresses": common.SchemaDataSourceIPAddresses("DSR LB"),
					"netmask":      common.SchemaDataSourceNetMask("DSR LB"),
					"gateway":      common.SchemaDataSourceGateway("DSR LB"),
					"vrid": schema.Int64Attribute{
						Computed:    true,
						Description: "The Virtual Router Identifier",
					},
				},
			},
			"vip": schema.ListNestedAttribute{
				Computed:    true,
				Description: "VIPs",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"vip": schema.StringAttribute{
							Computed:    true,
							Description: "The virtual IP address",
						},
						"port": schema.Int32Attribute{
							Computed:    true,
							Description: "The target port number for load-balancing",
						},
						"delay_loop": schema.Int32Attribute{
							Computed:    true,
							Description: "The interval in seconds between checks",
						},
						"sorry_server": schema.StringAttribute{
							Computed:    true,
							Description: "The IP address of the SorryServer. This will be used when all servers under this VIP are down",
						},
						"description": common.SchemaDataSourceDescription("DSR LB's VIP"),
						"server": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"ip_address": schema.StringAttribute{
										Computed:    true,
										Description: "The IP address of the destination server",
									},
									"protocol": schema.StringAttribute{
										Computed:    true,
										Description: desc.Sprintf("The protocol used for health checks. This will be one of [%s]", iaastypes.LoadBalancerHealthCheckProtocolStrings),
									},
									"path": schema.StringAttribute{
										Computed:    true,
										Description: "The path used when checking by HTTP/HTTPS",
									},
									"status": schema.Int32Attribute{
										Computed:    true,
										Description: "The response code to expect when checking by HTTP/HTTPS",
									},
									"retry": schema.Int32Attribute{
										Computed:    true,
										Description: "The retry count for server down detection, available only for TCP/HTTP/HTTPS",
									},
									"connect_timeout": schema.Int32Attribute{
										Computed:    true,
										Description: "The timeout in seconds for health checks, available only for TCP/HTTP/HTTPS",
									},
									"enabled": schema.BoolAttribute{
										Computed:    true,
										Description: "The flag to enable as destination of load balancing",
									},
								},
							},
						},
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing DSR LB (load_balancer in v2).",
	}
}

func (d *dsrLBDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dsrLBDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(data.Zone, d.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	searcher := iaas.NewLoadBalancerOp(d.client)
	res, err := searcher.Find(ctx, zone, common.CreateFindCondition(data.ID, data.Name, data.Tags))
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find DSR LB resource: %s", err))
		return
	}
	if res == nil || res.Count == 0 || len(res.LoadBalancers) == 0 {
		resp.Diagnostics.AddError("Read: Search Error", "no DSR LB found matching the given criteria")
		return
	}

	lb := res.LoadBalancers[0]
	if lb.Availability.IsFailed() {
		resp.Diagnostics.AddError("Read: State Error", fmt.Sprintf("got unexpected state: DSR LB[%s].Availability is failed", lb.ID.String()))
		return
	}
	data.updateState(lb, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
