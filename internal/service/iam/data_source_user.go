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
	"github.com/sacloud/iam-api-go/apis/user"
	v1 "github.com/sacloud/iam-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type userDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

func (d *userDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_user"
}

func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.IamClient
}

type userDataSourceModel struct {
	userBaseModel
}

func (d *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("IAM User"),
			"name":        common.SchemaDataSourceName("IAM User"),
			"description": common.SchemaDataSourceDescription("IAM User"),
			"code":        schemaDataSourceIAMCode("IAM User"),
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the IAM User",
			},
			"otp": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The OTP settings of the IAM User",
				Attributes: map[string]schema.Attribute{
					"status": schema.StringAttribute{
						Computed:    true,
						Description: "The OTP status of the IAM User",
					},
					"has_recovery_code": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the IAM User has recovery code for OTP",
					},
				},
			},
			"member": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The member information of the IAM User",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The member ID associated with the IAM User",
					},
					"code": schema.StringAttribute{
						Computed:    true,
						Description: "The member code associated with the IAM User",
					},
				},
			},
			"email": schema.StringAttribute{
				Computed:    true,
				Description: "The email of the IAM User",
			},
			"security_key_registered": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether a security key is registered for the IAM User",
			},
			"passwordless": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether passwordless authentication is enabled for the IAM User",
			},
			"created_at": common.SchemaDataSourceCreatedAt("IAM User"),
			"updated_at": common.SchemaDataSourceUpdatedAt("IAM User"),
		},
		MarkdownDescription: "Get information about an existing IAM User.",
	}
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data userDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var res *v1.User
	var err error
	userOp := iam.NewUserOp(d.client)
	if utils.IsKnown(data.Name) {
		perPage := 100 // TODO: Proper pagination if needed
		users, err := userOp.List(ctx, user.ListParams{PerPage: &perPage})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list IAM Users: %s", err))
			return
		}
		res, err = filterIAMUserByName(users.Items, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		res, err = userOp.Read(ctx, utils.MustAtoI(data.ID.ValueString()))
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read IAM User: %s", err))
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterIAMUserByName(keys []v1.User, name string) (*v1.User, error) {
	match := slices.Collect(func(yield func(v1.User) bool) {
		for _, v := range keys {
			if name != string(v.Name) {
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
		return nil, fmt.Errorf("multiple IAM User found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
