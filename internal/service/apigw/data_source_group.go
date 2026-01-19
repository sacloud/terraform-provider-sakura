// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type apigwGroupDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &apigwGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &apigwGroupDataSource{}
)

func NewApigwGroupDataSource() datasource.DataSource {
	return &apigwGroupDataSource{}
}

func (d *apigwGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_group"
}

func (d *apigwGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.ApigwClient
}

type apigwGroupDataSourceModel struct {
	apigwGroupBaseModel
}

func (d *apigwGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaDataSourceId("API Gateway Group"),
			"name": common.SchemaDataSourceName("API Gateway Group"),
			"tags": common.SchemaDataSourceComputedTags("API Gateway Group"),
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp",
			},
		},
		MarkdownDescription: "Get information about an existing API Gateway Group.",
	}
}

func (d *apigwGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data apigwGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupOp := apigw.NewGroupOp(d.client)
	var group *v1.Group
	var err error
	if utils.IsKnown(data.Name) {
		groups, err := groupOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list API Gateway groups: %s", err))
			return
		}
		group, err = filterAPIGWGroupByName(groups, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		group, err = groupOp.Read(ctx, uuid.MustParse(data.ID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read API Gateway group[%s]: %s", data.ID.ValueString(), err.Error()))
			return
		}
	}

	data.updateState(group)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterAPIGWGroupByName(keys []v1.Group, name string) (*v1.Group, error) {
	match := slices.Collect(func(yield func(v1.Group) bool) {
		for _, v := range keys {
			if name != string(v.Name.Value) {
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
		return nil, fmt.Errorf("multiple API Gateway groups found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
