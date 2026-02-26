// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	"github.com/sacloud/workflows-api-go"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowsDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &workflowsDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowsDataSource{}
)

func NewWorkflowsDataSource() datasource.DataSource {
	return &workflowsDataSource{}
}

func (d *workflowsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflows"
}

func (d *workflowsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.WorkflowsClient
}

type workflowsDataSourceModel struct {
	workflowBaseModel
}

func (d *workflowsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resourceName := "Workflows"

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId(resourceName),
			"name":        common.SchemaDataSourceName(resourceName),
			"description": common.SchemaDataSourceDescription(resourceName),
			"tags":        common.SchemaDataSourceTags(resourceName),
			"subscription_id": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The subscription ID of the %s.", resourceName),
			},
			"publish": schema.BoolAttribute{
				Computed:    true,
				Description: desc.Sprintf("Whether the %s is published.", resourceName),
			},
			"logging": schema.BoolAttribute{
				Computed:    true,
				Description: desc.Sprintf("Whether logging is enabled for the %s.", resourceName),
			},
			"service_principal_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The service principal id of the %s.", resourceName),
			},
			"concurrency_mode": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The concurrency mode of the %s.", resourceName),
			},
			"created_at": common.SchemaDataSourceCreatedAt(resourceName),
			"updated_at": common.SchemaDataSourceUpdatedAt(resourceName),
			"latest_revision": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "The ID of the revision.",
					},
					"runbook": schema.StringAttribute{
						Computed:    true,
						Description: "The runbook definition of the revision.",
					},
					"created_at": common.SchemaDataSourceCreatedAt(resourceName),
					"updated_at": common.SchemaDataSourceUpdatedAt(resourceName),
				},
			},
		},
		MarkdownDescription: "Get information about an existing Workflow.",
	}
}

func (d *workflowsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data workflowsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !utils.IsKnown(data.ID) {
		resp.Diagnostics.AddError("Read: Attribute Error", "'id' must be specified.")
		return
	}

	workflowOp := workflows.NewWorkflowOp(d.client)
	workflow, err := workflowOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read workflow[%s]: %s", data.ID.ValueString(), err))
		return
	}

	revisionOp := workflows.NewRevisionOp(d.client)
	revisions, err := revisionOp.List(ctx, v1.ListWorkflowRevisionsParams{ID: workflow.ID})
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list workflow revisions: %s", err))
		return
	}

	data.updateState(workflow)
	if err := data.updateRevisionsState(revisions.Revisions); err != nil {
		resp.Diagnostics.AddError("Read: State Error", fmt.Sprintf("failed to update revisions state: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
