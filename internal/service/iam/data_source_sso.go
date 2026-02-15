// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/iam-api-go"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type ssoDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &ssoDataSource{}
	_ datasource.DataSourceWithConfigure = &ssoDataSource{}
)

func NewSsoDataSource() datasource.DataSource {
	return &ssoDataSource{}
}

func (d *ssoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_sso"
}

func (d *ssoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type ssoDataSourceModel struct {
	ssoBaseModel
}

func (d *ssoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("IAM SSO"),
			"name":        common.SchemaDataSourceName("IAM SSO"),
			"description": common.SchemaDataSourceDescription("IAM SSO"),
			"idp_entity_id": schema.StringAttribute{
				Computed:    true,
				Description: "The IdP entity ID of the IAM SSO",
			},
			"idp_login_url": schema.StringAttribute{
				Computed:    true,
				Description: "The IdP login URL of the IAM SSO",
			},
			"idp_logout_url": schema.StringAttribute{
				Computed:    true,
				Description: "The IdP logout URL of the IAM SSO",
			},
			"idp_certificate": schema.StringAttribute{
				Computed:    true,
				Description: "The IdP certificate of the IAM SSO",
			},
			"sp_entity_id": schema.StringAttribute{
				Computed:    true,
				Description: "The SP entity ID of the IAM SSO",
			},
			"sp_acs_url": schema.StringAttribute{
				Computed:    true,
				Description: "The SP ACS URL of the IAM SSO",
			},
			"assigned": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the IAM SSO is assigned",
			},
			"created_at": common.SchemaDataSourceCreatedAt("IAM SSO"),
			"updated_at": common.SchemaDataSourceUpdatedAt("IAM SSO"),
		},
		MarkdownDescription: "Get information about an existing IAM SSO profile.",
	}
}

func (d *ssoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ssoDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *v1.SSOProfile
	var err error
	ssoOp := iam.NewSSOOp(d.client)
	if utils.IsKnown(data.Name) {
		perPage := 100 // TODO: Proper pagination if needed
		ssos, err := ssoOp.List(ctx, nil, &perPage)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list IAM SSO resources: %s", err))
			return
		}
		res, err = filterIAMSSOByName(ssos.Items, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		res, err = ssoOp.Read(ctx, utils.MustAtoI(data.ID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM SSO resource[%s]: %s", data.ID.ValueString(), err))
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterIAMSSOByName(keys []v1.SSOProfile, name string) (*v1.SSOProfile, error) {
	match := slices.Collect(func(yield func(v1.SSOProfile) bool) {
		for _, v := range keys {
			if name != v.Name {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple IAM SSO found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
