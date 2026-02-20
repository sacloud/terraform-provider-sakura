// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type authDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &authDataSource{}
	_ datasource.DataSourceWithConfigure = &authDataSource{}
)

func NewAuthDataSource() datasource.DataSource {
	return &authDataSource{}
}

func (d *authDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_auth"
}

func (d *authDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type authDataSourceModel struct {
	authBaseModel
}

func (d *authDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"password_policy": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"min_length": schema.Int32Attribute{
						Computed:    true,
						Description: "The minimum length of the password.",
					},
					"require_uppercase": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to require uppercase letters in the password.",
					},
					"require_lowercase": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to require lowercase letters in the password.",
					},
					"require_symbols": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to require symbols in the password.",
					},
				},
			},
			"conditions": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"ip_restriction": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"mode": schema.StringAttribute{
								Computed:    true,
								Description: "The mode of IP restriction.",
							},
							"source_network": schema.ListAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "The source networks for IP restriction.",
							},
						},
					},
					"require_two_factor_auth": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether to require two-factor authentication.",
					},
					"datetime_restriction": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"after": schema.StringAttribute{
								Computed:    true,
								Description: "The start time for datetime restriction.",
							},
							"before": schema.StringAttribute{
								Computed:    true,
								Description: "The end time for datetime restriction.",
							},
						},
					},
				},
			},
		},
		MarkdownDescription: "Get information about the IAM Authentication settings.",
	}
}

func (d *authDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data authDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ppRes, acRes := getAuth(ctx, d.client, &resp.State, &resp.Diagnostics)
	if ppRes == nil && acRes == nil {
		resp.Diagnostics.AddError("Read: API Error", "failed to read IAM Auth resource")
		return
	}

	data.updateState(ppRes, acRes)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
