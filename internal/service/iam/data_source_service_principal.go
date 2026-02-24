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
	"github.com/sacloud/iam-api-go/apis/serviceprincipal"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type servicePrincipalDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &servicePrincipalDataSource{}
	_ datasource.DataSourceWithConfigure = &servicePrincipalDataSource{}
)

func NewServicePrincipalDataSource() datasource.DataSource {
	return &servicePrincipalDataSource{}
}

func (d *servicePrincipalDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_service_principal"
}

func (d *servicePrincipalDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type servicePrincipalDataSourceModel struct {
	servicePrincipalBaseModel
}

func (d *servicePrincipalDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("IAM Service Principal"),
			"name":        common.SchemaDataSourceName("IAM Service Principal"),
			"description": common.SchemaDataSourceDescription("IAM Service Principal"),
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The project ID associated with the IAM Service Principal",
			},
			"created_at": common.SchemaDataSourceCreatedAt("IAM Service Principal"),
			"updated_at": common.SchemaDataSourceUpdatedAt("IAM Service Principal"),
		},
		MarkdownDescription: "Get information about an existing IAM Service Principal.",
	}
}

func (d *servicePrincipalDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data servicePrincipalDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *v1.ServicePrincipal
	var err error
	spOp := iam.NewServicePrincipalOp(d.client)
	if utils.IsKnown(data.Name) {
		perPage := 100 // TODO: Proper pagination if needed
		sps, err := spOp.List(ctx, serviceprincipal.ListParams{PerPage: &perPage})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list IAM Service Principal resources: %s", err))
			return
		}
		res, err = filterIAMServicePrincipalByName(sps.Items, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		res, err = spOp.Read(ctx, utils.MustAtoI(data.ID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM Service Principal resource[%s]: %s", data.ID.ValueString(), err))
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterIAMServicePrincipalByName(keys []v1.ServicePrincipal, name string) (*v1.ServicePrincipal, error) {
	match := slices.Collect(func(yield func(v1.ServicePrincipal) bool) {
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
		return nil, fmt.Errorf("multiple IAM Service Principals found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
