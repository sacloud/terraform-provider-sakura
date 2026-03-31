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

type lbnDataSource struct{ dataSourceClient }

type lbnDataSourceModel struct {
	lbNodeModel
	ClusterID          types.String `tfsdk:"cluster_id"`
	AutoScalingGroupID types.String `tfsdk:"auto_scaling_group_id"`
	LoadBalancerID     types.String `tfsdk:"lb_id"`
}

var (
	_ datasource.DataSource              = &lbnDataSource{}
	_ datasource.DataSourceWithConfigure = &lbnDataSource{}
)

func NewLoadBalancerNodeDataSource() datasource.DataSource {
	return &lbnDataSource{dataSourceNamed("lb_node")}
}

func (d *lbnDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "Information about an AppRun dedicated load balancer node",
		Attributes: map[string]schema.Attribute{
			"cluster_id":            d.schemaClusterID(),
			"auto_scaling_group_id": d.schemaASGID(),
			"lb_id": schema.StringAttribute{
				Required:    true,
				Description: "The load balancer ID that the lb_node belongs to",
				Validators:  []validator.String{sacloudvalidator.UUIDValidator},
			},
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
	}
}

func (d *lbnDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state lbnDataSourceModel
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

	nodeID, err := state.nodeID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Load Balancer Node ID", fmt.Sprintf("failed to parse load balancer node ID: %s", err))
		return
	}

	detail, err := d.api(cid, asgID).ReadNode(ctx, lbID, nodeID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated load balancer node: %s", err))
		return
	}

	if detail == nil {
		common.FilterNoResultErr(&res.Diagnostics)
		return
	}

	state.ClusterID = uuid2StringValue(cid)
	state.AutoScalingGroupID = uuid2StringValue(asgID)
	state.LoadBalancerID = uuid2StringValue(lbID)
	state.updateState(*detail)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (d *lbnDataSource) api(cid clusterID, asgID asgID) lb.LoadBalancerAPI {
	return lb.NewLoadBalancerOp(d.client, cid, asgID)
}

func (d *lbnDataSourceModel) clusterID() (clusterID, error) { return intoUUID[clusterID](d.ClusterID) }
func (d *lbnDataSourceModel) asgID() (asgID, error)         { return intoUUID[asgID](d.AutoScalingGroupID) }
func (d *lbnDataSourceModel) lbID() (lbID, error)           { return intoUUID[lbID](d.LoadBalancerID) }
func (d *lbnDataSourceModel) nodeID() (lbnID, error)        { return intoUUID[lbnID](d.ID) }
