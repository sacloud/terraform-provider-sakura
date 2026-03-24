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
	asg "github.com/sacloud/apprun-dedicated-api-go/apis/autoscalinggroup"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type asgDataSource struct{ dataSourceClient }
type asgDataSourceModel struct{ asgModel }

var (
	_ datasource.DataSource              = &asgDataSource{}
	_ datasource.DataSourceWithConfigure = &asgDataSource{}
)

func NewAutoScalingGroupDataSource() datasource.DataSource {
	return &asgDataSource{dataSourceNamed("auto_scaling_group")}
}

func (d *asgDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	id := d.schemaID()

	cid := d.schemaClusterID()

	name := d.schemaName()

	zone := schema.StringAttribute{
		Computed:    true,
		Description: "The zone name where the auto scaling group is located",
	}

	nameServers := schema.ListAttribute{
		Computed:    true,
		ElementType: types.StringType,
		Description: "The name servers for the auto scaling group",
	}

	workerServiceClassPath := schema.StringAttribute{
		Computed:    true,
		Description: "The worker service class path",
	}

	minNodes := schema.Int32Attribute{
		Computed:    true,
		Description: "Minimum number of nodes",
	}

	maxNodes := schema.Int32Attribute{
		Computed:    true,
		Description: "Maximum number of nodes",
	}

	currentNodes := schema.Int32Attribute{
		Computed:    true,
		Description: "The current number of nodes. You might want to ignore_changes this field because it changes from time to time",
	}

	deleting := schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the auto scaling group is being deleted",
	}

	ifaceIdx := schema.Int32Attribute{
		Computed:    true,
		Description: "The interface index",
	}

	upstream := schema.StringAttribute{
		Computed:    true,
		Description: "The upstream network",
	}

	start := schema.StringAttribute{
		Computed:    true,
		Description: "The start IP address of the range",
	}

	end := schema.StringAttribute{
		Computed:    true,
		Description: "The end IP address of the range",
	}

	ipRange := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"start": start,
			"end":   end,
		},
	}

	ipPool := schema.SetNestedAttribute{
		Computed:     true,
		NestedObject: ipRange,
		Description:  "The IP pool for the interface",
	}

	netmaskLen := schema.Int32Attribute{
		Computed:    true,
		Description: "The netmask length",
	}

	defaultGateway := schema.StringAttribute{
		Computed:    true,
		Description: "The default gateway",
	}

	packetFilterID := schema.StringAttribute{
		Computed:    true,
		Description: "The packet filter ID",
	}

	connectsToLB := schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the interface connects to the load balancer",
	}

	iface := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"interface_index":  ifaceIdx,
			"upstream":         upstream,
			"ip_pool":          ipPool,
			"netmask_len":      netmaskLen,
			"default_gateway":  defaultGateway,
			"packet_filter_id": packetFilterID,
			"connects_to_lb":   connectsToLB,
		},
	}

	interfaces := schema.SetNestedAttribute{
		Computed:     true,
		NestedObject: iface,
		Description:  "The network interfaces for the nodes",
	}

	res.Schema = schema.Schema{
		Description: "Information about an AppRun dedicated auto scaling group",
		Attributes: map[string]schema.Attribute{
			"id":                        id,
			"cluster_id":                cid,
			"name":                      name,
			"zone":                      zone,
			"name_servers":              nameServers,
			"worker_service_class_path": workerServiceClassPath,
			"min_nodes":                 minNodes,
			"max_nodes":                 maxNodes,
			"current_nodes":             currentNodes,
			"deleting":                  deleting,
			"interfaces":                interfaces,
		},
	}
}

func (d *asgDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state asgDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	var asgID *v1.AutoScalingGroupID
	var clusterID *v1.ClusterID
	var ds diag.Diagnostics

	if state.ID.IsNull() {
		clusterID, asgID, ds = state.byName(ctx, d)
	} else {
		clusterID, asgID, ds = state.byId(ctx, d)
	}
	res.Diagnostics.Append(ds...)

	if ds.HasError() {
		return
	}

	detail, err := d.api(*clusterID).Read(ctx, *asgID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated certificate: %s", err))
		return
	}

	if detail == nil {
		common.FilterNoResultErr(&res.Diagnostics)
		return
	}

	state.updateState(ctx, detail, *clusterID)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (state *asgDataSourceModel) byId(context.Context, *asgDataSource) (_ *v1.ClusterID, _ *v1.AutoScalingGroupID, d diag.Diagnostics) {
	a, e := state.asgID()

	if e != nil {
		d.AddError("Read: Invalid ID", fmt.Sprintf("failed to parse certificate ID: %s", e))
	}

	c, e := state.clusterID()

	if e != nil {
		d.AddError("Read: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", e))
	}

	return &c, &a, d
}

func (state *asgDataSourceModel) byName(ctx context.Context, r *asgDataSource) (_ *v1.ClusterID, _ *v1.AutoScalingGroupID, d diag.Diagnostics) {
	clusterID, err := state.clusterID()

	if err != nil {
		d.AddError("Read: Invalid Cluster ID", fmt.Sprintf("failed to parse certificate ID: %s", err))
		return
	}

	api := r.api(clusterID)
	certs, err := listed(func(cursor *v1.AutoScalingGroupID) ([]v1.ReadAutoScalingGroupDetail, *v1.AutoScalingGroupID, error) {
		return api.List(ctx, 10, cursor)
	})

	if err != nil {
		d.AddError("Read: API Error", fmt.Sprintf("failed to list AppRun Dedicated certificates: %s", err))
		return
	}

	name := state.Name.ValueString()
	for _, i := range certs {
		if i.Name == name {
			return &clusterID, &i.AutoScalingGroupID, d
		}
	}

	d.AddError("Read: API Error", fmt.Sprintf("certificate with name %q not found", name))
	return
}

func (d *asgDataSource) api(id v1.ClusterID) asg.AutoScalingGroupAPI {
	return asg.NewAutoScalingGroupOp(d.client, id)
}

func (s *asgDataSourceModel) clusterID() (v1.ClusterID, error) {
	return intoUUID[v1.ClusterID](s.ClusterID)
}

func (s *asgDataSourceModel) asgID() (v1.AutoScalingGroupID, error) {
	return intoUUID[v1.AutoScalingGroupID](s.ID)
}
