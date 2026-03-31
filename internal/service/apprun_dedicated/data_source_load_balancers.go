// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	lb "github.com/sacloud/apprun-dedicated-api-go/apis/loadbalancer"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
)

type loadBalancersDataSource struct{ dataSourceClient }

type loadBalancersDataSourceModel struct {
	ClusterID          types.String `tfsdk:"cluster_id"`
	AutoScalingGroupID types.String `tfsdk:"auto_scaling_group_id"`
	LoadBalancers      []lbModel    `tfsdk:"load_balancers"`
}

var (
	_ datasource.DataSource              = &loadBalancersDataSource{}
	_ datasource.DataSourceWithConfigure = &loadBalancersDataSource{}
)

func NewLoadBalancersDataSource() datasource.DataSource {
	return &loadBalancersDataSource{dataSourceNamed("load_balancers")}
}

func (d *loadBalancersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "List of load balancers in an AppRun dedicated auto scaling group",
		Attributes: map[string]schema.Attribute{
			"cluster_id":            d.schemaClusterID(),
			"auto_scaling_group_id": d.schemaASGID(),
			"load_balancers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of load balancers",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   d.schemaID(),
						"name": d.schemaName(),
						"service_class_path": schema.StringAttribute{
							Computed:    true,
							Description: "The service class path of the load balancer",
						},
						"name_servers": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "The name servers for the load balancer",
						},
						"interfaces": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"interface_index": schema.Int32Attribute{
										Computed:    true,
										Description: "The interface index",
									},
									"upstream": schema.StringAttribute{
										Computed:    true,
										Description: "The upstream network",
									},
									"ip_pool": schema.SetNestedAttribute{
										Computed: true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"start": schema.StringAttribute{
													Computed:    true,
													Description: "The start IP address of the range",
												},
												"end": schema.StringAttribute{
													Computed:    true,
													Description: "The end IP address of the range",
												},
											},
										},
										Description: "The IP pool for the interface",
									},
									"netmask": schema.Int32Attribute{
										Computed:    true,
										Description: "The netmask length",
									},
									"default_gateway": schema.StringAttribute{
										Computed:    true,
										Description: "The default gateway",
									},
									"vip": schema.StringAttribute{
										Computed:    true,
										Description: "The VIP address",
									},
									"virtual_router_id": schema.Int32Attribute{
										Computed:    true,
										Description: "The virtual router ID",
									},
									"packet_filter_id": schema.StringAttribute{
										Computed:    true,
										Description: "The packet filter ID",
									},
								},
							},
							Description: "The network interfaces for the load balancer",
						},
						"created": d.schemaCreatedAt(),
						"deleting": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the load balancer is being deleted",
						},
					},
				},
			},
		},
	}
}

func (d *loadBalancersDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state loadBalancersDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	cid, err := state.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	asgID, err := state.asgID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Auto Scaling Group ID", fmt.Sprintf("failed to parse auto scaling group ID: %s", err))
		return
	}

	api := d.api(cid, asgID)
	lbs, err := listed(func(cursor *lbID) ([]v1.ReadLoadBalancerSummary, *lbID, error) { return api.List(ctx, 10, cursor) })

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list load balancers: %s", err))
		return
	}

	state.ClusterID = uuid2StringValue(cid)
	state.AutoScalingGroupID = uuid2StringValue(asgID)

	// Fetch details for each load balancer
	state.LoadBalancers = make([]lbModel, 0, len(lbs))
	for _, summary := range lbs {
		detail, err := api.Read(ctx, summary.LoadBalancerID)

		if err != nil {
			res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read load balancer %s: %s", summary.LoadBalancerID, err))
			return
		}

		var model lbModel
		res.Diagnostics.Append(model.updateState(ctx, detail)...)

		if res.Diagnostics.HasError() {
			return
		}

		state.LoadBalancers = append(state.LoadBalancers, model)
	}

	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (d *loadBalancersDataSource) api(cid clusterID, asgID asgID) lb.LoadBalancerAPI {
	return lb.NewLoadBalancerOp(d.client, cid, asgID)
}

func (m *loadBalancersDataSourceModel) clusterID() (clusterID, error) {
	return intoUUID[clusterID](m.ClusterID)
}

func (m *loadBalancersDataSourceModel) asgID() (asgID, error) {
	return intoUUID[asgID](m.AutoScalingGroupID)
}
