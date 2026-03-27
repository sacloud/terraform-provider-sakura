// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	lb "github.com/sacloud/apprun-dedicated-api-go/apis/loadbalancer"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type lbnsDataSource struct{ dataSourceClient }

type lbnsDataSourceModel struct {
	ClusterID          types.String  `tfsdk:"cluster_id"`
	AutoScalingGroupID types.String  `tfsdk:"auto_scaling_group_id"`
	LoadBalancerID     types.String  `tfsdk:"load_balancer_id"`
	Nodes              []lbNodeModel `tfsdk:"nodes"`
}

var (
	_ datasource.DataSource              = &lbnsDataSource{}
	_ datasource.DataSourceWithConfigure = &lbnsDataSource{}
)

func NewLoadBalancerNodesDataSource() datasource.DataSource {
	return &lbnsDataSource{dataSourceNamed("load_balancer_nodes")}
}

func (d *lbnsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "List of load balancer nodes in an AppRun dedicated load balancer",
		Attributes: map[string]schema.Attribute{
			"cluster_id":            d.schemaClusterID(),
			"auto_scaling_group_id": d.schemaASGID(),
			"load_balancer_id": schema.StringAttribute{
				Required:    true,
				Description: "The load balancer ID that the load_balancer_nodes belong to",
				Validators:  []validator.String{sacloudvalidator.UUIDValidator},
			},
			"nodes": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of load balancer nodes",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "The load balancer node ID",
							Validators:  []validator.String{sacloudvalidator.UUIDValidator},
						},
						"resource_id": common.SchemaDataSourceId(d.name),
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "The status of the load balancer node",
						},
						"archive_version": schema.StringAttribute{
							Computed:    true,
							Description: "The archive version",
						},
						"create_error_message": schema.StringAttribute{
							Computed:    true,
							Description: "The error message if creation failed",
						},
						"created": schema.StringAttribute{
							Computed:    true,
							Description: "The creation time of the load balancer node",
						},
						"interfaces": schema.ListNestedAttribute{
							Computed:    true,
							Description: "The network interfaces of the load balancer node",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"interface_index": schema.Int32Attribute{
										Computed:    true,
										Description: "The interface index",
									},
									"addresses": schema.SetNestedAttribute{
										Computed:    true,
										Description: "The IP addresses assigned to this interface",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"address": schema.StringAttribute{
													Computed:    true,
													Description: "The IP address",
												},
												"vip": schema.BoolAttribute{
													Computed:    true,
													Description: "Whether this is a VIP address",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *lbnsDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state lbnsDataSourceModel
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

	lbID, err := state.lbID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Load Balancer ID", fmt.Sprintf("failed to parse load balancer ID: %s", err))
		return
	}

	api := d.api(cid, asgID)
	list, err := api.ListNodes(ctx, lbID, 10, nil)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list load balancer nodes: %s", err))
		return
	}

	state.ClusterID = uuid2StringValue(cid)
	state.AutoScalingGroupID = uuid2StringValue(asgID)
	state.LoadBalancerID = uuid2StringValue(lbID)

	// Fetch details for each node
	state.Nodes = make([]lbNodeModel, 0, len(list))
	for _, summary := range list {
		detail, err := api.ReadNode(ctx, lbID, summary.LoadBalancerNodeID)

		if err != nil {
			res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read load balancer node %s: %s", summary.LoadBalancerNodeID, err))
			return
		}

		var model lbNodeModel
		model.updateState(*detail)
		state.Nodes = append(state.Nodes, model)
	}

	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (d *lbnsDataSource) api(cid clusterID, asgID asgID) lb.LoadBalancerAPI {
	return lb.NewLoadBalancerOp(d.client, cid, asgID)
}

func (m *lbnsDataSourceModel) clusterID() (clusterID, error) { return intoUUID[clusterID](m.ClusterID) }
func (m *lbnsDataSourceModel) asgID() (asgID, error)         { return intoUUID[asgID](m.AutoScalingGroupID) }
func (m *lbnsDataSourceModel) lbID() (lbID, error)           { return intoUUID[lbID](m.LoadBalancerID) }
