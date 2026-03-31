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

type lbsDataSource struct{ dataSourceClient }
type lbsDataSourceModel struct{ lbServiceClassModel }

var (
	_ datasource.DataSource              = &lbsDataSource{}
	_ datasource.DataSourceWithConfigure = &lbsDataSource{}
)

func NewLoadBalancerServiceClassDataSource() datasource.DataSource {
	return &lbsDataSource{dataSourceNamed("lb_service_class")}
}

func (d *lbsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "Information about a specific load balancer service class for AppRun Dedicated",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the load balancer service class",
			},
			"path": schema.StringAttribute{
				Computed:    true,
				Description: "The service class path",
			},
			"node_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of nodes assigned",
			},
		},
	}
}

func (d *lbsDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state lbsDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	classes, err := service_class.NewServiceClassOp(d.client).ListLB(ctx)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list load balancer service classes: %s", err))
		return
	}

	targetName := state.Name.ValueString()
	var found *v1.ReadLbServiceClass
	for _, c := range classes {
		if c.Name == targetName {
			found = &c
			break
		}
	}

	if found == nil {
		res.Diagnostics.AddError(
			"Read: Not Found",
			fmt.Sprintf("load balancer service class with name %q not found", targetName),
		)
		return
	}

	state.updateState(*found)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}
