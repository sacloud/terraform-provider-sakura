// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	"github.com/sacloud/workflows-api-go"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type subscriptionDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &subscriptionDataSource{}
	_ datasource.DataSourceWithConfigure = &subscriptionDataSource{}
)

func NewSubscriptionDataSource() datasource.DataSource {
	return &subscriptionDataSource{}
}

func (d *subscriptionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflows_subscription"
}

func (d *subscriptionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.WorkflowsClient
}

type workflowsSubscriptionDataSourceModel struct {
	workflowsSubscriptionBaseModel
}

func (d *subscriptionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resourceName := "Workflows Subscription"

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The ID of the %s.", resourceName),
			},
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The account ID of the %s.", resourceName),
			},
			"contract_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The contract ID of the %s.", resourceName),
			},
			"plan_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The plan ID of the %s.", resourceName),
			},
			"plan_name": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The plan name of the %s.", resourceName),
			},
			"activate_from": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The activate from timestamp of the %s.", resourceName),
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The creation timestamp of the %s.", resourceName),
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The last update timestamp of the %s.", resourceName),
			},
		},
		MarkdownDescription: "Get information about a current Workflow Subscription.",
	}
}

func (d *subscriptionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data workflowsSubscriptionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	subscriptionOp := workflows.NewSubscriptionOp(d.client)
	subscription, err := subscriptionOp.Read(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Workflows Subscription: %s", err))
		return
	}

	data.updateState(subscription)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
