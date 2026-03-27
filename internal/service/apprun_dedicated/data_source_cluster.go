// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/sacloud/apprun-dedicated-api-go/apis/cluster"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type clusterDataSource struct{ dataSourceClient }
type clusterDataSourceModel struct{ clusterModel }

var (
	_ datasource.DataSource              = &clusterDataSource{}
	_ datasource.DataSourceWithConfigure = &clusterDataSource{}
)

func NewClusterDataSource() datasource.DataSource {
	return &clusterDataSource{dataSourceNamed("cluster")}
}

func (d *clusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "Information about an AppRun dedicated cluster",
		Attributes: map[string]schema.Attribute{
			"id":   d.schemaID(),
			"name": d.schemaName(),
			"service_principal_id": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal ID.  This is the principal that invokes the application",
			},
			"ports": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of ports that the cluster listens on (max 5)",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port": schema.Int32Attribute{
							Computed:    true,
							Description: "The port number where the cluster listens for requests",
						},
						"protocol": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Either `http`, `https`, or `tcp`",
						},
					},
				},
			},
			"has_lets_encrypt_email": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "If true the cluster must listen HTTP port 80 because LetsEncrypt challenges there",
			},
			"created_at": d.schemaCreatedAt(),
		},
	}
}

func (d *clusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state clusterDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	var id *clusterID
	var ds diag.Diagnostics

	if state.ID.IsNull() {
		id, ds = state.byName(ctx, d)
	} else {
		id, ds = state.byId(ctx, d)
	}
	res.Diagnostics.Append(ds...)

	if id == nil {
		return
	}

	if ds.HasError() {
		return
	}

	detail, err := d.api().Read(ctx, *id)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated cluster: %s", err))
		return
	}

	if detail == nil {
		common.FilterNoResultErr(&res.Diagnostics)
		return
	}

	state.updateState(detail)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (state *clusterDataSourceModel) byId(context.Context, *clusterDataSource) (ret *clusterID, d diag.Diagnostics) {
	id, err := state.clusterID()

	if err != nil {
		d.AddError("Read: Invalid ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	ret = &id
	return
}

func (state *clusterDataSourceModel) byName(ctx context.Context, d *clusterDataSource) (ret *clusterID, ds diag.Diagnostics) {
	list, err := listed(func(c *clusterID) ([]cluster.ClusterDetail, *clusterID, error) { return d.api().List(ctx, 10, c) })

	if err != nil {
		ds.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated cluster: %s", err))
		return
	}

	name := state.Name.ValueString()
	for _, i := range list {
		if i.Name == name {
			ret = &i.ClusterID
			return
		}
	}

	ds.AddError("Read: API Error", fmt.Sprintf("cluster with name %q not found", name))
	return
}

func (r *clusterDataSource) api() *cluster.ClusterOp { return cluster.NewClusterOp(r.client) }
