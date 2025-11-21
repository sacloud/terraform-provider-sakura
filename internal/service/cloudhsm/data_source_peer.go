// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cloudhsm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/cloudhsm-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type cloudHSMPeerDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &cloudHSMPeerDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudHSMPeerDataSource{}
)

func NewCloudHSMPeerDataSource() datasource.DataSource {
	return &cloudHSMPeerDataSource{}
}

func (d *cloudHSMPeerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudhsm_peer"
}

func (d *cloudHSMPeerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.CloudHSMClient
}

type cloudHSMPeerDataSourceModel struct {
	cloudHSMPeerBaseModel
}

func (d *cloudHSMPeerDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaDataSourceId("CloudHSM Peer"),
			"cloudhsm_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the CloudHSM to associate with the client",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"index": schema.Int64Attribute{
				Computed:    true,
				Description: "The index of the CloudHSM Peer",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the CloudHSM Peer",
			},
			"routes": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The routes for the CloudHSM Peer",
			},
		},
		MarkdownDescription: "Get information about an existing CloudHSM Peer.",
	}
}

func (d *cloudHSMPeerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudHSMPeerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if len(id) == 0 {
		resp.Diagnostics.AddError("Missing Attribute", "'id' must be specified.")
		return
	}

	chsm := getCloudHSM(ctx, d.client, data.CloudHSMID.ValueString(), &resp.State, &resp.Diagnostics)
	if chsm == nil {
		return
	}

	chsmPeer := getCloudHSMPeer(ctx, d.client, chsm, id, &resp.State, &resp.Diagnostics)
	if chsmPeer == nil {
		return
	}

	data.updateState(chsmPeer, chsm.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
