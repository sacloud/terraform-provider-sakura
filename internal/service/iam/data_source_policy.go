// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type policyDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &policyDataSource{}
	_ datasource.DataSourceWithConfigure = &policyDataSource{}
)

func NewPolicyDataSource() datasource.DataSource {
	return &policyDataSource{}
}

func (d *policyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_policy"
}

func (d *policyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type policyDataSourceModel struct {
	policyBaseModel
}

func (d *policyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"target": schema.StringAttribute{
				Required:    true,
				Description: "The target of the IAM Policy",
				Validators: []validator.String{
					stringvalidator.OneOf("project", "folder", "organization"),
				},
			},
			"target_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the target. Required for Folder or Project",
			},
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
		MarkdownDescription: "Get information about an existing IAM Policy.",
	}
}

func (d *policyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data policyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res := getIAMPolicy(ctx, d.client, data.Target.ValueString(), data.TargetID.ValueString(), &resp.Diagnostics)
	if res == nil {
		return
	}

	data.updateState(data.Target.ValueString(), res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
