// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	lb "github.com/sacloud/apprun-dedicated-api-go/apis/loadbalancer"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type lbDataSource struct{ dataSourceClient }
type lbDataSourceModel struct {
	lbModel
	ClusterID          types.String `tfsdk:"cluster_id"`
	AutoScalingGroupID types.String `tfsdk:"auto_scaling_group_id"`
}

var (
	_ datasource.DataSource              = &lbDataSource{}
	_ datasource.DataSourceWithConfigure = &lbDataSource{}
)

func NewLoadBalancerDataSource() datasource.DataSource {
	return &lbDataSource{dataSourceNamed("lb")}
}

func (d *lbDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "Information about an AppRun dedicated load balancer",
		Attributes: map[string]schema.Attribute{
			"id":                    d.schemaID(),
			"cluster_id":            d.schemaClusterID(),
			"auto_scaling_group_id": d.schemaASGID(),
			"name":                  d.schemaName(),
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
	}
}

func (d *lbDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state lbDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	var lbID *lbID
	var clusterID *clusterID
	var asgID *asgID
	var ds diag.Diagnostics

	if state.ID.IsNull() {
		clusterID, asgID, lbID, ds = state.byName(ctx, d)
	} else {
		clusterID, asgID, lbID, ds = state.byId(ctx, d)
	}
	res.Diagnostics.Append(ds...)

	if ds.HasError() {
		return
	}

	detail, err := d.api(*clusterID, *asgID).Read(ctx, *lbID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated load balancer: %s", err))
		return
	}

	if detail == nil {
		common.FilterNoResultErr(&res.Diagnostics)
		return
	}

	state.ClusterID = uuid2StringValue(*clusterID)
	state.AutoScalingGroupID = uuid2StringValue(*asgID)
	res.Diagnostics.Append(state.updateState(ctx, detail)...)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (state *lbDataSourceModel) byId(context.Context, *lbDataSource) (_ *clusterID, _ *asgID, _ *lbID, d diag.Diagnostics) {
	l, e := state.lbID()

	if e != nil {
		d.AddError("Read: Invalid ID", fmt.Sprintf("failed to parse load balancer ID: %s", e))
	}

	c, e := state.clusterID()

	if e != nil {
		d.AddError("Read: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", e))
	}

	a, e := state.asgID()

	if e != nil {
		d.AddError("Read: Invalid Auto Scaling Group ID", fmt.Sprintf("failed to parse auto scaling group ID: %s", e))
	}

	return &c, &a, &l, d
}

func (state *lbDataSourceModel) byName(ctx context.Context, r *lbDataSource) (_ *clusterID, _ *asgID, _ *lbID, d diag.Diagnostics) {
	clusterID, err := state.clusterID()

	if err != nil {
		d.AddError("Read: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	asgID, err := state.asgID()

	if err != nil {
		d.AddError("Read: Invalid Auto Scaling Group ID", fmt.Sprintf("failed to parse auto scaling group ID: %s", err))
		return
	}

	api := r.api(clusterID, asgID)
	lbs, err := listed(func(cursor *lbID) ([]v1.ReadLoadBalancerSummary, *lbID, error) { return api.List(ctx, 10, cursor) })

	if err != nil {
		d.AddError("Read: API Error", fmt.Sprintf("failed to list AppRun Dedicated load balancers: %s", err))
		return
	}

	name := state.Name.ValueString()
	for _, i := range lbs {
		if i.Name == name {
			return &clusterID, &asgID, &i.LoadBalancerID, d
		}
	}

	d.AddError("Read: API Error", fmt.Sprintf("load balancer with name %q not found", name))
	return
}

func (d *lbDataSource) api(cid clusterID, asgID asgID) lb.LoadBalancerAPI {
	return lb.NewLoadBalancerOp(d.client, cid, asgID)
}

func (s *lbDataSourceModel) clusterID() (clusterID, error) { return intoUUID[clusterID](s.ClusterID) }
func (s *lbDataSourceModel) asgID() (asgID, error)         { return intoUUID[asgID](s.AutoScalingGroupID) }
