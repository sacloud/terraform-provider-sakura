// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package nlb

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

type nlbDataSource struct {
	client *common.APIClient
}

func NewNLBDataSource() datasource.DataSource {
	return &nlbDataSource{}
}

var (
	_ datasource.DataSource              = &nlbDataSource{}
	_ datasource.DataSourceWithConfigure = &nlbDataSource{}
)

func (r *nlbDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nlb"
}

func (r *nlbDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type nlbDataSourceModel struct {
	nlbBaseModel
}

func (r *nlbDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("NLB"),
			"name":        common.SchemaDataSourceName("NLB"),
			"description": common.SchemaDataSourceDescription("NLB"),
			"tags":        common.SchemaDataSourceTags("NLB"),
			"zone":        common.SchemaDataSourceZone("NLB"),
			"icon_id":     common.SchemaDataSourceIconID("NLB"),
			"plan":        common.SchemaDataSourcePlan("NLB", []string{"standard", "highspec"}),
			"network_interface": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Network interface for NLB",
				Attributes: map[string]schema.Attribute{
					"vswitch_id":   common.SchemaDataSourceVSwitchID("NLB"),
					"ip_addresses": common.SchemaDataSourceIPAddresses("NLB"),
					"netmask":      common.SchemaDataSourceNetMask("NLB"),
					"gateway":      common.SchemaDataSourceGateway("NLB"),
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
						"description": common.SchemaDataSourceDescription("NLB's VIP"),
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
		MarkdownDescription: "Get information about an existing NLB.",
	}
}

func (d *nlbDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data nlbDataSourceModel
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
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find NLB resource: %s", err))
		return
	}
	if res == nil || res.Count == 0 || len(res.LoadBalancers) == 0 {
		resp.Diagnostics.AddError("Read: Search Error", "no NLB found matching the given criteria")
		return
	}

	lb := res.LoadBalancers[0]
	if lb.Availability.IsFailed() {
		resp.Diagnostics.AddError("Read: State Error", fmt.Sprintf("got unexpected state: NLB[%s].Availability is failed", lb.ID.String()))
		return
	}
	data.updateState(lb, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
