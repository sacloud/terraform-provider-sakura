// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type lbServiceClassModel struct {
	Path      types.String `tfsdk:"path"`
	Name      types.String `tfsdk:"name"`
	NodeCount types.Int32  `tfsdk:"node_count"`
}

type workerServiceClassModel struct {
	Path types.String `tfsdk:"path"`
	Name types.String `tfsdk:"name"`
}

func (m *lbServiceClassModel) updateState(class v1.ReadLbServiceClass) {
	m.Path = types.StringValue(class.Path)
	m.Name = types.StringValue(class.Name)
	m.NodeCount = types.Int32Value(common.ToInt32(class.NodeCount))
}

func (m *workerServiceClassModel) updateState(class v1.ReadWorkerServiceClass) {
	m.Path = types.StringValue(class.Path)
	m.Name = types.StringValue(class.Name)
}
