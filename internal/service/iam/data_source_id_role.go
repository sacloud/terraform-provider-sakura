// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iam-api-go"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type idRoleDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &idRoleDataSource{}
	_ datasource.DataSourceWithConfigure = &idRoleDataSource{}
)

func NewIdRoleDataSource() datasource.DataSource {
	return &idRoleDataSource{}
}

func (d *idRoleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_id_role"
}

func (d *idRoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type idRoleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (d *idRoleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("IAM ID Role"),
			"name":        common.SchemaDataSourceName("IAM ID Role"),
			"description": common.SchemaDataSourceDescription("IAM ID Role"),
		},
	}
}

func (d *idRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data idRoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleOp := iam.NewIDRoleOp(d.client)
	res, err := roleOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM ID Role resource: %s", err))
		return
	}

	data.ID = types.StringValue(res.ID)
	data.Name = types.StringValue(res.Name)
	data.Description = types.StringValue(res.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
