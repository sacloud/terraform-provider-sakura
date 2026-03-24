// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/apprun-dedicated-api-go/apis/service_class"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type wscDataSource struct{ dataSourceClient }
type wscDataSourceModel struct {
	Classes []workerServiceClassModel `tfsdk:"classes"`
}

var (
	_ datasource.DataSource              = &wscDataSource{}
	_ datasource.DataSourceWithConfigure = &wscDataSource{}
)

func NewWorkerServiceClassesDataSource() datasource.DataSource {
	return &wscDataSource{dataSourceNamed("worker_service_classes")}
}

func (d *wscDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	name := schema.StringAttribute{
		Computed:    true,
		Description: "The service class name",
	}

	path := schema.StringAttribute{
		Computed:    true,
		Description: "The service class path",
	}

	class := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"path": path,
			"name": name,
		},
	}

	res.Schema = schema.Schema{
		Description: "List of available worker service classes for AppRun Dedicated",
		Attributes: map[string]schema.Attribute{
			"classes": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "List of worker service classes",
				NestedObject: class,
			},
		},
	}
}

func (d *wscDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state wscDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	classes, err := service_class.NewServiceClassOp(d.client).ListWorker(ctx)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list worker service classes: %s", err))
		return
	}

	state.Classes = common.MapTo(classes, func(src v1.ReadWorkerServiceClass) (dst workerServiceClassModel) {
		dst.updateState(src)
		return
	})

	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}
