// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	seccon "github.com/sacloud/security-control-api-go"
	v1 "github.com/sacloud/security-control-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type activationDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &activationDataSource{}
	_ datasource.DataSourceWithConfigure = &activationDataSource{}
)

func NewActivationDataSource() datasource.DataSource {
	return &activationDataSource{}
}

func (d *activationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_control_activation"
}

func (d *activationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.SecurityControlClient
}

type activationDataSourceModel struct {
	activationBaseModel
}

func (d *activationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_principal_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Service Principal ID associated with the Project",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the Security Control is enabled",
			},
			"automated_action_limit": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of registerable automated actions",
			},
		},
		MarkdownDescription: "Get information about an existing Security Control Activation.",
	}
}

func (d *activationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data activationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	actOp := seccon.NewActivationOp(d.client)
	result, err := actOp.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Security Control Activation status: %s", err))
		return
	}

	data.updateState(result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
