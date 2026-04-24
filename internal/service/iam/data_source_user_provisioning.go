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
	"github.com/sacloud/iam-api-go/apis/scim"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type userProvisioningDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &userProvisioningDataSource{}
	_ datasource.DataSourceWithConfigure = &userProvisioningDataSource{}
)

func NewUserProvisioningDataSource() datasource.DataSource {
	return &userProvisioningDataSource{}
}

func (d *userProvisioningDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_user_provisioning"
}

func (d *userProvisioningDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type userProvisioningDataSourceModel struct {
	userProvisioningBaseModel
}

func (d *userProvisioningDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaDataSourceId("IAM User Provisioning"),
			"name": common.SchemaDataSourceName("IAM User Provisioning"),
			"base_url": schema.StringAttribute{
				Computed:    true,
				Description: "The base URL of the IAM User Provisioning SCIM endpoint.",
			},
			"created_at": common.SchemaDataSourceCreatedAt("IAM User Provisioning"),
			"updated_at": common.SchemaDataSourceUpdatedAt("IAM User Provisioning"),
		},
		MarkdownDescription: "Get information about an existing IAM User Provisioning.",
	}
}

func (d *userProvisioningDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userProvisioningDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *v1.ScimConfigurationBase
	var err error
	scimOp := iam.NewScimOp(d.client)
	if utils.IsKnown(data.Name) {
		perPage := 100 // TODO: Proper pagination if needed
		scims, err := scimOp.List(ctx, scim.ListParams{PerPage: &perPage})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list IAM User Provisioning: %s", err))
			return
		}
		res, err = filterIAMScimByName(scims.Items, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		res, err = scimOp.Read(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM User Provisioning[%s]: %s", data.ID.ValueString(), err))
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterIAMScimByName(keys []v1.ScimConfigurationBase, name string) (*v1.ScimConfigurationBase, error) {
	match := slices.Collect(func(yield func(v1.ScimConfigurationBase) bool) {
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
		return nil, fmt.Errorf("multiple IAM User Provisionings found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
