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
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	wn "github.com/sacloud/apprun-dedicated-api-go/apis/workernode"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type wnDataSource struct{ dataSourceClient }

type wnDataSourceModel struct {
	wnModel
	ClusterID          types.String `tfsdk:"cluster_id"`
	AutoScalingGroupID types.String `tfsdk:"auto_scaling_group_id"`
}

var (
	_ datasource.DataSource              = &wnDataSource{}
	_ datasource.DataSourceWithConfigure = &wnDataSource{}
)

func NewWorkerNodeDataSource() datasource.DataSource {
	return &wnDataSource{dataSourceNamed("worker_node")}
}

func (d *wnDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
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

	res.Schema = schema.Schema{
		Description: "Information about an AppRun dedicated worker node",
		Attributes: map[string]schema.Attribute{
			"cluster_id":            cid,
			"auto_scaling_group_id": aid,
			"id":                    id,
			"resource_id":           resourceID,
			"draining":              draining,
			"status":                status,
			"healthy":               healthy,
			"creating":              creating,
			"created":               created,
			"running_containers":    runningContainers,
			"network_interfaces":    networkInterfaces,
			"archive_version":       archiveVersion,
			"create_error_message":  createErrorMessage,
		},
	}
}

func (d *wnDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state wnDataSourceModel
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

	wnID, err := state.wnID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Worker Node ID", fmt.Sprintf("failed to parse worker node ID: %s", err))
		return
	}

	detail, err := d.api(cid, asgID).Read(ctx, wnID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated worker node: %s", err))
		return
	}

	if detail == nil {
		common.FilterNoResultErr(&res.Diagnostics)
		return
	}

	state.ClusterID = uuid2StringValue(cid)
	state.AutoScalingGroupID = uuid2StringValue(asgID)
	state.updateState(ctx, detail)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (d *wnDataSource) api(cid v1.ClusterID, asgID v1.AutoScalingGroupID) wn.WorkerNodeAPI {
	return wn.NewWorkerNodeOp(d.client, cid, asgID)
}

func (d *wnDataSourceModel) asgID() (v1.AutoScalingGroupID, error) {
	return intoUUID[v1.AutoScalingGroupID](d.AutoScalingGroupID)
}

func (d *wnDataSourceModel) clusterID() (v1.ClusterID, error) {
	return intoUUID[v1.ClusterID](d.ClusterID)
}
