// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/apprun-dedicated-api-go/apis/cluster"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type clusterDataSource struct{ client *v1.Client }
type clusterDataSourceModel struct{ clusterModel }

var (
	_ datasource.DataSource              = &clusterDataSource{}
	_ datasource.DataSourceWithConfigure = &clusterDataSource{}
)

func NewClusterDataSource() datasource.DataSource { return new(clusterDataSource) }

func (*clusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, res *datasource.MetadataResponse) {
	res.TypeName = req.ProviderTypeName + "_apprun_dedicated_cluster"
}

func (d *clusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, res *datasource.ConfigureResponse) {
	client := common.GetApiClientFromProvider(req.ProviderData, &res.Diagnostics)

	if client == nil {
		return
	}

	d.client = client.AppRunDedicatedClient
}

func (*clusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	id := common.SchemaDataSourceId("cluster").(schema.StringAttribute)
	id.Required = false
	id.Optional = true
	id.Computed = true
	id.Validators = []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
		sacloudvalidator.UUIDValidator,
	}

	name := common.SchemaDataSourceName("cluster").(schema.StringAttribute)
	name.Required = false
	name.Optional = true
	name.Computed = true
	id.Validators = []validator.String{
		stringvalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}

	spid := schema.StringAttribute{
		Computed:    true,
		Description: "The service principal ID.  This is the principal that invokes the application",
	}

	portno := schema.Int32Attribute{
		Computed:    true,
		Description: "The port number where the cluster listens for requests",
	}

	protocol := schema.StringAttribute{
		Computed:            true,
		MarkdownDescription: "Either `http`, `https`, or `tcp`",
	}

	nested := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"port":     portno,
			"protocol": protocol,
		},
	}

	ports := schema.ListNestedAttribute{
		Computed:     true,
		Description:  "The list of ports that the cluster listens on (max 5)",
		NestedObject: nested,
	}

	le := schema.BoolAttribute{
		Computed:            true,
		MarkdownDescription: "If true the cluster must listen HTTP port 80 because LetsEncrypt challenges there",
	}

	at := common.SchemaDataSourceCreatedAt("cluster")

	cluster := schema.Schema{
		Description: "Information about an AppRun dedicated cluster",
		Attributes: map[string]schema.Attribute{
			"id":                     id,
			"name":                   name,
			"service_principal_id":   spid,
			"ports":                  ports,
			"has_lets_encrypt_email": le,
			"created_at":             at,
		},
	}

	res.Schema = cluster
}

func (d *clusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state clusterDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	var id *v1.ClusterID

	if state.ID.IsNull() {
		id = d.byName(ctx, req, res, &state)
	} else {
		id = d.byId(ctx, req, res, &state)
	}

	if id == nil {
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

func (d *clusterDataSource) byId(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse, state *clusterDataSourceModel) *v1.ClusterID {
	id, err := state.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return nil
	}

	return &id
}

func (d *clusterDataSource) byName(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse, state *clusterDataSourceModel) *v1.ClusterID {
	list, err := listed(func(c *v1.ClusterID) ([]cluster.ClusterDetail, *v1.ClusterID, error) { return d.api().List(ctx, 10, c) })

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated cluster: %s", err))
		return nil
	}

	name := state.Name.ValueString()
	for _, i := range list {
		if i.Name == name {
			return &i.ClusterID
		}
	}

	res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("cluster with name %q not found", name))
	return nil
}

func (r *clusterDataSource) api() *cluster.ClusterOp { return cluster.NewClusterOp(r.client) }
