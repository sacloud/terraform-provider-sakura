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
	wn "github.com/sacloud/apprun-dedicated-api-go/apis/workernode"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type workerNodesDataSource struct{ dataSourceClient }

type wnsDataSourceModel struct {
	ClusterID          types.String `tfsdk:"cluster_id"`
	AutoScalingGroupID types.String `tfsdk:"auto_scaling_group_id"`
	Nodes              []wnModel    `tfsdk:"nodes"`
}

var (
	_ datasource.DataSource              = &workerNodesDataSource{}
	_ datasource.DataSourceWithConfigure = &workerNodesDataSource{}
)

func NewWorkerNodesDataSource() datasource.DataSource {
	return &workerNodesDataSource{dataSourceNamed("worker_nodes")}
}

func (d *workerNodesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	cid := d.schemaClusterID()

	aid := d.schemaASGID()

	id := d.schemaID()

	resourceID := common.SchemaDataSourceId(d.name)

	draining := schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the worker node is draining",
	}

	status := schema.StringAttribute{
		Computed:    true,
		Description: "The status of the worker node",
	}

	healthy := schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the worker node is healthy",
	}

	creating := schema.BoolAttribute{
		Computed:    true,
		Description: "Whether the worker node is being created",
	}

	created := schema.StringAttribute{
		Computed:    true,
		Description: "The creation time of the worker node",
	}

	containerID := schema.StringAttribute{
		Computed:    true,
		Description: "The container ID",
	}

	containerName := schema.StringAttribute{
		Computed:    true,
		Description: "The container name",
	}

	state := schema.StringAttribute{
		Computed:    true,
		Description: "The container state",
	}

	containerStatus := schema.StringAttribute{
		Computed:    true,
		Description: "The container status",
	}

	image := schema.StringAttribute{
		Computed:    true,
		Description: "The container image",
	}

	startedAt := schema.StringAttribute{
		Computed:    true,
		Description: "The container start time",
	}

	applicationID := schema.StringAttribute{
		Computed:    true,
		Description: "The ID of the application that the worker node belongs to",
		Validators:  []validator.String{sacloudvalidator.UUIDValidator},
	}

	applicationVersion := schema.Int32Attribute{
		Computed:    true,
		Description: "The application's version",
	}

	runningContainer := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"container_id":        containerID,
			"name":                containerName,
			"state":               state,
			"status":              containerStatus,
			"image":               image,
			"started_at":          startedAt,
			"application_id":      applicationID,
			"application_version": applicationVersion,
		},
	}

	runningContainers := schema.ListNestedAttribute{
		Computed:     true,
		Description:  "The list of running containers on the worker node",
		NestedObject: runningContainer,
	}

	interfaceIndex := schema.Int32Attribute{
		Computed:    true,
		Description: "The interface index",
	}

	addresses := schema.SetAttribute{
		Computed:    true,
		ElementType: types.StringType,
		Description: "The IP address(es) assigned to this interface",
	}

	networkInterface := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"interface_index": interfaceIndex,
			"addresses":       addresses,
		},
	}

	networkInterfaces := schema.ListNestedAttribute{
		Computed:     true,
		Description:  "The network interfaces of the worker node",
		NestedObject: networkInterface,
	}

	archiveVersion := schema.StringAttribute{
		Computed:    true,
		Description: "The archive version",
	}

	createErrorMessage := schema.StringAttribute{
		Computed:    true,
		Description: "The error message if creation failed",
	}

	workerNode := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id":                   id,
			"resource_id":          resourceID,
			"draining":             draining,
			"status":               status,
			"healthy":              healthy,
			"creating":             creating,
			"created":              created,
			"running_containers":   runningContainers,
			"network_interfaces":   networkInterfaces,
			"archive_version":      archiveVersion,
			"create_error_message": createErrorMessage,
		},
	}

	workerNodes := schema.ListNestedAttribute{
		Computed:     true,
		Description:  "List of worker nodes",
		NestedObject: workerNode,
	}

	res.Schema = schema.Schema{
		Description: "List of worker nodes in an AppRun dedicated auto scaling group",
		Attributes: map[string]schema.Attribute{
			"cluster_id":            cid,
			"auto_scaling_group_id": aid,
			"nodes":                 workerNodes,
		},
	}
}

func (d *workerNodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state wnsDataSourceModel
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
	nodes, err := listed(func(cursor *wnID) ([]wn.WorkerNodeDetail, *wnID, error) { return api.List(ctx, 10, cursor) })

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list worker nodes: %s", err))
		return
	}

	state.ClusterID = uuid2StringValue(cid)
	state.AutoScalingGroupID = uuid2StringValue(asgID)
	state.Nodes = common.MapTo(nodes, func(src wn.WorkerNodeDetail) (dst wnModel) {
		res.Diagnostics.Append(dst.updateState(ctx, &src)...)
		return
	})

	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (d *workerNodesDataSource) api(cid clusterID, asgID asgID) wn.WorkerNodeAPI {
	return wn.NewWorkerNodeOp(d.client, cid, asgID)
}

func (m *wnsDataSourceModel) clusterID() (clusterID, error) { return intoUUID[clusterID](m.ClusterID) }
func (m *wnsDataSourceModel) asgID() (asgID, error)         { return intoUUID[asgID](m.AutoScalingGroupID) }
