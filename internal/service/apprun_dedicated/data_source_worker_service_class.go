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
)

type wsDataSource struct{ dataSourceClient }
type wsDataSourceModel struct{ workerServiceClassModel }

var (
	_ datasource.DataSource              = &wsDataSource{}
	_ datasource.DataSourceWithConfigure = &wsDataSource{}
)

func NewWorkerServiceClassDataSource() datasource.DataSource {
	return &wsDataSource{dataSourceNamed("worker_service_class")}
}

func (d *wsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "Information about a specific worker service class for AppRun Dedicated",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the worker service class",
			},
			"path": schema.StringAttribute{
				Computed:    true,
				Description: "The service class path",
			},
		},
	}
}

func (d *wsDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state wsDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	classes, err := service_class.NewServiceClassOp(d.client).ListWorker(ctx)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list worker service classes: %s", err))
		return
	}

	targetName := state.Name.ValueString()
	var found *v1.ReadWorkerServiceClass
	for _, c := range classes {
		if c.Name == targetName {
			found = &c
			break
		}
	}

	if found == nil {
		res.Diagnostics.AddError(
			"Read: Not Found",
			fmt.Sprintf("worker service class with name %q not found", targetName),
		)
		return
	}

	state.updateState(*found)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}
