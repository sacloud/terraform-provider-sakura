// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/iam-api-go"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type orgIDPolicyDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &orgIDPolicyDataSource{}
	_ datasource.DataSourceWithConfigure = &orgIDPolicyDataSource{}
)

func NewOrgIDPolicyDataSource() datasource.DataSource {
	return &orgIDPolicyDataSource{}
}

func (d *orgIDPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_organization_id_policy"
}

func (d *orgIDPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type orgIDPolicyDataSourceModel struct {
	idPolicyBaseModel
}

func (d *orgIDPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"bindings": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The bindings of the IAM Organization ID Policy",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"role": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "The role of the IAM Organization ID Policy",
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Computed:    true,
									Description: "The type of the role",
								},
								"id": schema.StringAttribute{
									Computed:    true,
									Description: "The ID of the IAM Organization ID Policy",
								},
							},
						},
						"principals": schema.ListNestedAttribute{
							Computed:    true,
							Description: "The principals of the IAM Organization ID Policy",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Computed:    true,
										Description: "The type of the principal",
									},
									"id": schema.StringAttribute{
										Computed:    true,
										Description: "The ID of the principal",
									},
								},
							},
						},
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing IAM Organization ID Policy.",
	}
}

func (d *orgIDPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data orgIDPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	op := iam.NewIDPolicyOp(d.client)
	res, err := op.ReadOrganizationIdPolicy(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM Organization ID Policy resource: %s", err))
		return
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
