// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package security_control

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	seccon "github.com/sacloud/security-control-api-go"
	v1 "github.com/sacloud/security-control-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type automatedActionDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &automatedActionDataSource{}
	_ datasource.DataSourceWithConfigure = &automatedActionDataSource{}
)

func NewAutomatedActionDataSource() datasource.DataSource {
	return &automatedActionDataSource{}
}

func (d *automatedActionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_control_automated_action"
}

func (d *automatedActionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.SecurityControlClient
}

type automatedActionDataSourceModel struct {
	automatedActionBaseModel
}

func (d *automatedActionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("Automated Action"),
			"name":        common.SchemaDataSourceName("Automated Action"),
			"description": common.SchemaDataSourceDescription("Automated Action"),
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the Automated Action is enabled",
			},
			"action": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The settings for Automated Action",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:    true,
						Description: "The triggered type of Automated Action",
					},
					"parameters": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "The parameters for Automated Action",
						Attributes: map[string]schema.Attribute{
							"service_principal_id": schema.StringAttribute{
								Computed:    true,
								Description: "The Service Principal ID associated with the Automated Action",
							},
							"target_id": schema.StringAttribute{
								Computed:    true,
								Description: "The id of target resource for the Automated Action",
							},
							"revision": schema.Int64Attribute{
								Computed:    true,
								Description: "The revision number of workflow to be executed",
							},
							"revision_alias": schema.StringAttribute{
								Computed:    true,
								Description: "The revision alias of workflow to be executed",
							},
							"args": schema.StringAttribute{
								Computed:    true,
								Description: "The arguments to be passed to the workflow",
							},
							"name": schema.StringAttribute{
								Computed:    true,
								Description: "The name of the workflow execution",
							},
						},
					},
				},
			},
			"execution_condition": schema.StringAttribute{
				Computed:    true,
				Description: "The CEL expression that defines the condition when Automated Action is triggered",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation timestamp of the Automated Action",
			},
		},
		MarkdownDescription: "Get information about an existing Security Control's Automated Action.",
	}
}

func (d *automatedActionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data automatedActionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aaOp := seccon.NewAutomatedActionsOp(d.client)
	var res *v1.AutomatedActionOutput
	var err error
	if utils.IsKnown(data.Name) {
		aas, err := aaOp.List(ctx, v1.AutomatedActionsListParams{PageSize: v1.NewOptInt(100)})
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list Security Control's Automated Action resources: %s", err))
			return
		}
		res, err = filterAutomatedActionByName(aas.Items, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: Search Error", err.Error())
			return
		}
	} else {
		res, err = aaOp.Read(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read Security Control's Automated Action[%s] resource: %s", data.ID.ValueString(), err))
			return
		}
	}

	data.updateState(res)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterAutomatedActionByName(aas []v1.AutomatedActionOutput, name string) (*v1.AutomatedActionOutput, error) {
	match := slices.Collect(func(yield func(v1.AutomatedActionOutput) bool) {
		for _, v := range aas {
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
		return nil, fmt.Errorf("multiple Automated Action resources found with the same condition. name=%q", name)
	}

	return &match[0], nil
}
