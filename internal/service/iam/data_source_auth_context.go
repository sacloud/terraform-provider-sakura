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
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type authContextDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &authContextDataSource{}
	_ datasource.DataSourceWithConfigure = &authContextDataSource{}
)

func NewAuthContextDataSource() datasource.DataSource {
	return &authContextDataSource{}
}

func (d *authContextDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_auth_context"
}

func (d *authContextDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type authContextDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	AuthType           types.String `tfsdk:"auth_type"`
	LimitedToProjectID types.String `tfsdk:"limited_to_project_id"`
}

func (d *authContextDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID of the API Key or Service Principal.",
			},
			"auth_type": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The authentication type. This will be one of [%s].", common.MapTo(v1.GetAuthContextOKAuthTypeApikey.AllValues(), common.ToString)),
			},
			"limited_to_project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The operable project ID by the API Key or Service Principal",
			},
		},
		MarkdownDescription: "Get information about an existing IAM Auth Context.",
	}
}

func (d *authContextDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data authContextDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	authOp := iam.NewAuthOp(d.client)
	res, err := authOp.ReadAuthContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM Auth Context: %s", err))
		return
	}

	data.ID = types.StringValue(utils.ItoA(res.ResourceID))
	data.AuthType = types.StringValue(string(res.AuthType))
	if res.LimitedToProjectID.IsNull() {
		data.LimitedToProjectID = types.StringNull()
	} else {
		data.LimitedToProjectID = types.StringValue(utils.ItoA(res.LimitedToProjectID.Value))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
