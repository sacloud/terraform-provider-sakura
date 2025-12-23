// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cloudhsm

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/cloudhsm-api-go"
	v1 "github.com/sacloud/cloudhsm-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type cloudHSMClientDataSource struct {
	client *common.APIClient
}

var (
	_ datasource.DataSource              = &cloudHSMClientDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudHSMClientDataSource{}
)

func NewCloudHSMClientDataSource() datasource.DataSource {
	return &cloudHSMClientDataSource{}
}

func (d *cloudHSMClientDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudhsm_client"
}

func (d *cloudHSMClientDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient
}

type cloudHSMClientDataSourceModel struct {
	cloudHSMClientBaseModel
}

func (d *cloudHSMClientDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaDataSourceId("CloudHSM Client"),
			"name": common.SchemaDataSourceName("CloudHSM Client"),
			"zone": schemaDataSourceZone("CloudHSM Client"),
			"cloudhsm_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the CloudHSM to associate with the client",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"certificate": schema.StringAttribute{
				Computed:    true,
				Description: "The certificate for the CloudHSM Client",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation date of the CloudHSM Client",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "The modification date of the CloudHSM Client",
			},
			"availability": schema.StringAttribute{
				Computed:    true,
				Description: "The availability status of the CloudHSM Client",
			},
		},
		MarkdownDescription: "Get information about an existing CloudHSM Client.",
	}
}

func (d *cloudHSMClientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudHSMClientDataSourceModel
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
	chsm := getCloudHSM(ctx, client, data.CloudHSMID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsm == nil {
		return
	}

	clientOp, err := cloudhsm.NewClientOp(client, chsm)
	if err != nil {
		resp.Diagnostics.AddError("Read: Client Error", fmt.Sprintf("failed to create CloudHSM Client operation: %s", err))
		return
	}

	var chsmClient *v1.CloudHSMClient
	if len(name) > 0 {
		chsmClients, err := clientOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list CloudHSM Client resource: %s", err))
			return
		}
		chsmClient, err = FilterCloudHsmClientByName(chsmClients, name)
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", fmt.Sprintf("failed to filter CloudHSM Client resource by name: %s", err.Error()))
			return
		}
	} else {
		chsmClient, err = clientOp.Read(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to find CloudHSM Client resource[%s]: %s", id, err.Error()))
			return
		}
	}

	data.updateState(chsmClient, zone, chsm.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FilterCloudHsmClientByName(cloudHSMs []v1.CloudHSMClient, name string) (*v1.CloudHSMClient, error) {
	match := slices.Collect(func(yield func(v1.CloudHSMClient) bool) {
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
		return nil, fmt.Errorf("multiple CloudHSM Client resources found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
