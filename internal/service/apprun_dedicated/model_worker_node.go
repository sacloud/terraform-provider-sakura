// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	wn "github.com/sacloud/apprun-dedicated-api-go/apis/workernode"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type wnID = v1.WorkerNodeID

type wnifModel struct {
	Index     types.Int32 `tfsdk:"interface_index"`
	Addresses types.Set   `tfsdk:"addresses"`
}

type containerModel struct {
	ContainerID        types.String `tfsdk:"container_id"`
	Name               types.String `tfsdk:"name"`
	State              types.String `tfsdk:"state"`
	Status             types.String `tfsdk:"status"`
	Image              types.String `tfsdk:"image"`
	StartedAt          types.String `tfsdk:"started_at"`
	ApplicationID      types.String `tfsdk:"application_id"`
	ApplicationVersion types.Int32  `tfsdk:"application_version"`
}

type wnModel struct {
	ID                 types.String     `tfsdk:"id"`
	ResourceID         types.String     `tfsdk:"resource_id"`
	Draining           types.Bool       `tfsdk:"draining"`
	Status             types.String     `tfsdk:"status"`
	Healthy            types.Bool       `tfsdk:"healthy"`
	Creating           types.Bool       `tfsdk:"creating"`
	Created            types.String     `tfsdk:"created"`
	RunningContainers  []containerModel `tfsdk:"running_containers"`
	NetworkInterfaces  []wnifModel      `tfsdk:"network_interfaces"`
	ArchiveVersion     types.String     `tfsdk:"archive_version"`
	CreateErrorMessage types.String     `tfsdk:"create_error_message"`
}

var wnifAttrs = attrTypes{
	"interface_index": types.Int32Type,
	"addresses":       types.SetType{ElemType: types.StringType},
}

var containerAttrs = attrTypes{
	"container_id":        types.StringType,
	"name":                types.StringType,
	"state":               types.StringType,
	"status":              types.StringType,
	"image":               types.StringType,
	"started_at":          types.StringType,
	"application_id":      types.StringType,
	"application_version": types.Int32Type,
}

var wnAttrs = attrTypes{
	"cluster_id":            types.StringType,
	"auto_scaling_group_id": types.StringType,
	"id":                    types.StringType,
	"resource_id":           types.StringType,
	"draining":              types.BoolType,
	"status":                types.StringType,
	"healthy":               types.BoolType,
	"creating":              types.BoolType,
	"created":               types.StringType,
	"running_containers":    types.SetType{ElemType: types.ObjectType{AttrTypes: containerAttrs}},
	"network_interfaces":    types.SetType{ElemType: types.ObjectType{AttrTypes: wnifAttrs}},
	"archive_version":       types.StringType,
	"create_error_message":  types.StringType,
}

func (wnifModel) AttributeTypes() attrTypes      { return wnifAttrs }
func (containerModel) AttributeTypes() attrTypes { return containerAttrs }
func (wnModel) AttributeTypes() attrTypes        { return wnAttrs }

func (i *wnifModel) updateState(ctx context.Context, ni wn.WorkerNodeNetworkInterface) (ret diag.Diagnostics) {
	i.Index = types.Int32Value(common.ToInt32(ni.InterfaceIndex))
	i.Addresses, ret = types.SetValueFrom(ctx, types.StringType, ni.Addresses)
	return
}

func (r *containerModel) updateState(rc v1.RunningContainer) {
	r.ContainerID = types.StringValue(rc.ContainerID)
	r.Name = types.StringValue(rc.Name)
	r.State = types.StringValue(rc.State)
	r.Status = types.StringValue(rc.Status)
	r.Image = types.StringValue(rc.Image)
	r.StartedAt = intoRFC2822(rc.StartedAt)
	r.ApplicationID = uuid2StringValue(rc.ApplicationID)
	r.ApplicationVersion = types.Int32Value(rc.ApplicationVersion)
	return
}

func (m *wnModel) updateState(ctx context.Context, detail *wn.WorkerNodeDetail) (ret diag.Diagnostics) {
	m.ID = uuid2StringValue(detail.WorkerNodeID)
	m.ResourceID = types.StringPointerValue(detail.ResourceID)
	m.Draining = types.BoolValue(detail.Draining)
	m.Status = types.StringValue(string(detail.Status))
	m.Healthy = types.BoolValue(detail.Healthy)
	m.Creating = types.BoolValue(detail.Creating)
	m.Created = intoRFC2822(detail.Created)
	m.ArchiveVersion = types.StringPointerValue(detail.ArchiveVersion)
	m.CreateErrorMessage = types.StringPointerValue(detail.CreateErrorMessage)
	m.RunningContainers = common.MapTo(detail.RunningContainers, stateUpdater[v1.RunningContainer, containerModel])
	m.NetworkInterfaces = common.MapTo(detail.NetworkInterfaces, func(src wn.WorkerNodeNetworkInterface) (dst wnifModel) {
		ret.Append(dst.updateState(ctx, src)...)
		return
	})
	return
}
