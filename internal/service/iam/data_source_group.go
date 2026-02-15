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
	"github.com/sacloud/iam-api-go/apis/group"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type groupDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

func (d *groupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_group"
}

func (d *groupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type groupDataSourceModel struct {
	groupBaseModel
}

func (d *groupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("IAM Group"),
			"name":        common.SchemaDataSourceName("IAM Group"),
			"description": common.SchemaDataSourceDescription("IAM Group"),
			"created_at":  common.SchemaDataSourceCreatedAt("IAM Group"),
			"updated_at":  common.SchemaDataSourceUpdatedAt("IAM Group"),
		},
		MarkdownDescription: "Get information about an existing IAM Group.",
	}
}

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data groupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *v1.Group
	var err error
	groupOp := iam.NewGroupOp(d.client)
	if utils.IsKnown(data.Name) {
		perPage := 100 // TODO: Proper pagination if needed
		groups, err := groupOp.List(ctx, group.ListParams{PerPage: &perPage})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list IAM Group resources: %s", err))
			return
		}
		res, err = filterIAMGroupByName(groups.Items, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		res, err = groupOp.Read(ctx, utils.MustAtoI(data.ID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM Group resource[%s]: %s", data.ID.ValueString(), err))
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterIAMGroupByName(keys []v1.Group, name string) (*v1.Group, error) {
	match := slices.Collect(func(yield func(v1.Group) bool) {
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
		return nil, fmt.Errorf("multiple IAM Groups found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
