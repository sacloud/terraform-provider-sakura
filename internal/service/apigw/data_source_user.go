// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apigw

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/apigw-api-go"
	v1 "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type apigwUserDataSource struct {
	client *v1.Client
}

func NewApigwUserDataSource() datasource.DataSource {
	return &apigwUserDataSource{}
}

var (
	_ datasource.DataSource              = &apigwUserDataSource{}
	_ datasource.DataSourceWithConfigure = &apigwUserDataSource{}
)

func (r *apigwUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apigw_user"
}

func (r *apigwUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.ApigwClient
}

type apigwUserDataSourceModel struct {
	apigwUserBaseModel
	Authentication *apigwUserAuthenticationModel `tfsdk:"authentication"`
}

func (r *apigwUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":         common.SchemaDataSourceId("API Gateway User"),
			"name":       common.SchemaDataSourceName("API Gateway User"),
			"tags":       common.SchemaDataSourceComputedTags("API Gateway User"),
			"created_at": schemaDataSourceAPIGWCreatedAt("API Gateway User"),
			"updated_at": schemaDataSourceAPIGWUpdatedAt("API Gateway User"),
			"custom_id": schema.StringAttribute{
				Computed:    true,
				Description: "The custom ID of the API Gateway User",
			},
			"ip_restriction": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "IP restriction configuration for the user",
				Attributes: map[string]schema.Attribute{
					"protocols": schema.StringAttribute{
						Computed:    true,
						Description: "The protocols to restrict by",
					},
					"restricted_by": schema.StringAttribute{
						Computed:    true,
						Description: "The category to restrict by",
					},
					"ips": schema.SetAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "The IPv4 addresses to be restricted",
					},
				},
			},
			"groups": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Groups associated with the user",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "ID of the API Gateway Group",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the API Gateway Group",
						},
					},
				},
			},
			"authentication": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Authentication information of the API Gateway User",
				Attributes: map[string]schema.Attribute{
					"basic_auth": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Computed:    true,
								Description: "The basic auth username",
							},
						},
					},
					"jwt": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"key": schema.StringAttribute{
								Computed:    true,
								Description: "The JWT key",
							},
							"algorithm": schema.StringAttribute{
								Computed:    true,
								Description: "The JWT algorithm",
							},
						},
					},
					"hmac_auth": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"username": schema.StringAttribute{
								Computed:    true,
								Description: "The HMAC auth username",
							},
						},
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing API Gateway User.",
	}
}

func (d *apigwUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data apigwUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userOp := apigw.NewUserOp(d.client)
	var id string
	if utils.IsKnown(data.Name) {
		users, err := userOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list API Gateway users: %s", err))
			return
		}
		user, err := filterAPIGWUserByName(users, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
		id = user.ID.Value.String()
	} else {
		id = data.ID.ValueString()
	}

	user := getAPIGWUser(ctx, d.client, id, &resp.State, &resp.Diagnostics)
	if user == nil {
		return
	}

	ueOp := apigw.NewUserExtraOp(d.client, user.ID.Value)
	auth, err := ueOp.ReadAuth(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read authentication settings for API Gateway User[%s]: %s", user.ID.Value.String(), err))
		return
	}

	data.updateState(user)
	data.Authentication = flattenAPIGWUserAuthentication(auth)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterAPIGWUserByName(keys []v1.User, name string) (*v1.User, error) {
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
		return nil, fmt.Errorf("multiple API Gateway services found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
