// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cloudhsm

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/cloudhsm-api-go"
	v1 "github.com/sacloud/cloudhsm-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type cloudHSMDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &cloudHSMDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudHSMDataSource{}
)

func NewCloudHSMDataSource() datasource.DataSource {
	return &cloudHSMDataSource{}
}

func (d *cloudHSMDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudhsm"
}

func (d *cloudHSMDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type cloudHSMDataSourceModel struct {
	cloudHSMBaseModel
}

func (d *cloudHSMDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("CloudHSM"),
			"description": common.SchemaDataSourceDescription("CloudHSM"),
			"tags":        common.SchemaDataSourceTags("CloudHSM"),
			"name":        common.SchemaDataSourceName("CloudHSM"),
			"zone":        schemaDataSourceZone("CloudHSM"),
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation date of the CloudHSM",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "The modification date of the CloudHSM",
			},
			"availability": schema.StringAttribute{
				Computed:    true,
				Description: "The availability status of the CloudHSM",
			},
			"ipv4_network_address": schema.StringAttribute{
				Computed:    true,
				Description: "The IPv4 network address of the CloudHSM",
			},
			"ipv4_netmask": schema.Int32Attribute{
				Computed:    true,
				Description: "The IPv4 netmask of the CloudHSM",
			},
			"ipv4_address": schema.StringAttribute{
				Computed:    true,
				Description: "The IPv4 address of the CloudHSM",
			},
			"local_router": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The local router information of the CloudHSM",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The ID of the local router",
					},
					"secret_key": schema.StringAttribute{
						Computed:    true,
						Sensitive:   true,
						Description: "The secret key of the local router",
					},
				},
			},
		},
		MarkdownDescription: "Get information about an existing CloudHSM.",
	}
}

func (d *cloudHSMDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudHSMDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	name := data.Name.ValueString()
	if len(name) == 0 && len(id) == 0 {
		resp.Diagnostics.AddError("Read: Attribute Error", "either 'id' or 'name' must be specified.")
		return
	}

	zone := getZone(data.Zone, d.client, &resp.Diagnostics)
	client := createClient(zone, d.client)
	cloudhsmOp := cloudhsm.NewCloudHSMOp(client)
	var cloudhsm *v1.CloudHSM
	var err error
	if len(name) > 0 {
		cloudHSMs, err := cloudhsmOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find CloudHSM resource: %s", err))
			return
		}
		cloudhsm, err = FilterCloudHsmByName(cloudHSMs, name)
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", fmt.Sprintf("failed to filter CloudHSM resource by name: %s", err))
			return
		}
	} else {
		cloudhsm, err = cloudhsmOp.Read(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find CloudHSM resource[%s]: %s", id, err))
			return
		}
	}

	data.updateState(cloudhsm, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FilterCloudHsmByName(cloudHSMs []v1.CloudHSM, name string) (*v1.CloudHSM, error) {
	match := slices.Collect(func(yield func(v1.CloudHSM) bool) {
		for _, v := range cloudHSMs {
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
		return nil, fmt.Errorf("multiple CloudHSM resources found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
