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

type lbscDataSource struct{ dataSourceClient }
type lbscDataSourceModel struct {
	Classes []lbServiceClassModel `tfsdk:"classes"`
}

var (
	_ datasource.DataSource              = &lbscDataSource{}
	_ datasource.DataSourceWithConfigure = &lbscDataSource{}
)

func NewLoadBalancerServiceClassesDataSource() datasource.DataSource {
	return &lbscDataSource{dataSourceNamed("load_balancer_service_classes")}
}

func (d *lbscDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	name := schema.StringAttribute{
		Computed:    true,
		Description: "The service class name",
	}

	path := schema.StringAttribute{
		Computed:    true,
		Description: "The service class path",
	}

	count := schema.Int32Attribute{
		Computed:    true,
		Description: "The number of nodes assigned",
	}

	class := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"path":       path,
			"name":       name,
			"node_count": count,
		},
	}

	res.Schema = schema.Schema{
		Description: "List of available load balancer service classes for AppRun Dedicated",
		Attributes: map[string]schema.Attribute{
			"classes": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "List of load balancer service classes",
				NestedObject: class,
			},
		},
	}
}

func (d *lbscDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state lbscDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	classes, err := service_class.NewServiceClassOp(d.client).ListLB(ctx)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list load balancer service classes: %s", err))
		return
	}

	state.Classes = common.MapTo(classes, stateUpdater[v1.ReadLbServiceClass, lbServiceClassModel])

	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}
