// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package workflows

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
	"github.com/sacloud/workflows-api-go"
	v1 "github.com/sacloud/workflows-api-go/apis/v1"
)

type workflowRevisionAliasDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &workflowRevisionAliasDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowRevisionAliasDataSource{}
)

func NewWorkflowsRevisionAliasDataSource() datasource.DataSource {
	return &workflowRevisionAliasDataSource{}
}

func (d *workflowRevisionAliasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflows_revision_alias"
}

func (d *workflowRevisionAliasDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.WorkflowsClient
}

type workflowRevisionAliasDataSourceModel struct {
	workflowRevisionAliasBaseModel
}

func (d *workflowRevisionAliasDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resourceName := "Workflows RevisionAlias"

	resp.Schema = datasourceSchema.Schema{
		Attributes: map[string]datasourceSchema.Attribute{
			"workflow_id": datasourceSchema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The workflow ID of the %s.", resourceName),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"revision_id": datasourceSchema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The revision ID of the %s.", resourceName),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					sacloudvalidator.StringFuncValidator(func(value string) error {
						_, err := strconv.Atoi(value)
						return err
					}),
				},
			},
			"alias": datasourceSchema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The alias name of the %s.", resourceName),
			},
		},
		MarkdownDescription: "Get information about an existing Workflows Revision Alias.",
	}
}

func (d *workflowRevisionAliasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data workflowRevisionAliasDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// validatorで数値であることは保証されているため、エラーは発生しない前提
	revisionID := utils.MustAtoI(data.RevisionID.ValueString())

	revisionOp := workflows.NewRevisionOp(d.client)
	rev, err := revisionOp.Read(ctx, data.WorkflowID.ValueString(), revisionID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Read: API Error",
			fmt.Sprintf("failed to read workflow revision: %s", err),
		)
		return
	}

	updateRevisionAliasState(&data.workflowRevisionAliasBaseModel, rev)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
