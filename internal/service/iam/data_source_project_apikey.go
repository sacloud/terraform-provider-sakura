// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iam-api-go"
	"github.com/sacloud/iam-api-go/apis/projectapikey"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type projectApiKeyDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &projectApiKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &projectApiKeyDataSource{}
)

func NewProjectApiKeyDataSource() datasource.DataSource {
	return &projectApiKeyDataSource{}
}

func (d *projectApiKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_project_apikey"
}

func (d *projectApiKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type projectApiKeyDataSourceModel struct {
	projectApiKeyBaseModel
}

func (d *projectApiKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("IAM Project API Key"),
			"name":        common.SchemaDataSourceName("IAM Project API Key"),
			"description": common.SchemaDataSourceDescription("IAM Project API Key"),
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The project ID associated with the IAM Project API Key",
			},
			"iam_roles": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The IAM roles assigned to the IAM Project API Key",
			},
			"server_resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "The server resource ID of IAM Project API Key.",
			},
			"zone": schema.StringAttribute{
				Computed:    true,
				Description: "The zone of IAM Project API Key.",
			},
			"access_token": schema.StringAttribute{
				Computed:    true,
				Description: "The access token of the IAM Project API Key.",
			},
			"created_at": common.SchemaDataSourceCreatedAt("IAM Project API Key"),
			"updated_at": common.SchemaDataSourceUpdatedAt("IAM Project API Key"),
		},
	}
}

func (d *projectApiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectApiKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *v1.ProjectApiKey
	var err error
	paKeyOp := iam.NewProjectAPIKeyOp(d.client)
	if utils.IsKnown(data.Name) {
		perPage := 100 // TODO: Proper pagination if needed
		paKeys, err := paKeyOp.List(ctx, projectapikey.ListParams{PerPage: &perPage})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list IAM Project API Key resources: %s", err))
			return
		}
		res, err = filterIAMProjectApiKeyByName(paKeys.Items, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		res, err = paKeyOp.Read(ctx, utils.MustAtoI(data.ID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM Project API Key resource: %s", err))
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterIAMProjectApiKeyByName(keys []v1.ProjectApiKey, name string) (*v1.ProjectApiKey, error) {
	match := slices.Collect(func(yield func(v1.ProjectApiKey) bool) {
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
		return nil, fmt.Errorf("multiple IAM Project API Keys found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
