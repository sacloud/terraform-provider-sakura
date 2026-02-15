// Copyright 2016-2026 terraform-provider-sakura authors
// SPDX-License-Identifier: Apache-2.0

package iam

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/iam-api-go"
	"github.com/sacloud/iam-api-go/apis/project"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type projectDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

func (d *projectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_project"
}

func (d *projectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type projectDataSourceModel struct {
	projectBaseModel
}

func (d *projectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("IAM Project"),
			"name":        common.SchemaDataSourceName("IAM Project"),
			"description": common.SchemaDataSourceDescription("IAM Project"),
			"code": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The code of the IAM Project",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the IAM Project",
			},
			"parent_folder_id": schema.StringAttribute{
				Computed:    true,
				Description: "The parent folder ID associated with the IAM Project",
			},
			"created_at": common.SchemaDataSourceCreatedAt("IAM Project"),
			"updated_at": common.SchemaDataSourceUpdatedAt("IAM Project"),
		},
		MarkdownDescription: "Get information about an existing IAM Project.",
	}
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *v1.Project
	var err error
	projectOp := iam.NewProjectOp(d.client)
	if utils.IsKnown(data.Name) || utils.IsKnown(data.Code) {
		perPage := 100 // TODO: Proper pagination if needed
		projects, err := projectOp.List(ctx, project.ListParams{PerPage: &perPage})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list IAM Project resources: %s", err))
			return
		}
		if utils.IsKnown(data.Code) {
			res, err = filterIAMProjectByCode(projects.Items, data.Code.ValueString())
		} else {
			res, err = filterIAMProjectByName(projects.Items, data.Name.ValueString())
		}
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		res, err = projectOp.Read(ctx, utils.MustAtoI(data.ID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM Project resource[%s]: %s", data.ID.ValueString(), err))
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterIAMProjectByName(keys []v1.Project, name string) (*v1.Project, error) {
	match := slices.Collect(func(yield func(v1.Project) bool) {
		for _, v := range keys {
			if !strings.Contains(v.Name, name) {
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
		return nil, fmt.Errorf("multiple IAM Project found with the same condition. name=%q", name)
	}
	return &match[0], nil
}

func filterIAMProjectByCode(keys []v1.Project, code string) (*v1.Project, error) {
	match := slices.Collect(func(yield func(v1.Project) bool) {
		for _, v := range keys {
			if code != v.Code {
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
		return nil, fmt.Errorf("multiple IAM Project found with the same condition. code=%q", code)
	}
	return &match[0], nil
}
