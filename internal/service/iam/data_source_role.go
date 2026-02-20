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

type roleDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &roleDataSource{}
	_ datasource.DataSourceWithConfigure = &roleDataSource{}
)

func NewRoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

func (d *roleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_role"
}

func (d *roleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type roleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Category    types.String `tfsdk:"category"`
}

func (d *roleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("IAM Role"),
			"name":        common.SchemaDataSourceName("IAM Role"),
			"description": common.SchemaDataSourceDescription("IAM Role"),
			"category": schema.StringAttribute{
				Computed:    true,
				Description: "The category of the IAM Role",
			},
		},
		MarkdownDescription: "Get information about an existing IAM Role.",
	}
}

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data roleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleOp := iam.NewIAMRoleOp(d.client)
	res, err := roleOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM Role resource[%s]: %s", data.ID.ValueString(), err))
		return
	}

	data.ID = types.StringValue(res.ID)
	data.Name = types.StringValue(res.Name)
	data.Description = types.StringValue(res.Description)
	data.Category = types.StringValue(res.Category)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
