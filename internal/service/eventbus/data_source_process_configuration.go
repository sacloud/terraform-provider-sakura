// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/eventbus-api-go"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type processConfigurationDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &processConfigurationDataSource{}
	_ datasource.DataSourceWithConfigure = &processConfigurationDataSource{}
)

func NewEventBusProcessConfigurationDataSource() datasource.DataSource {
	return &processConfigurationDataSource{}
}

func (d *processConfigurationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eventbus_process_configuration"
}

func (d *processConfigurationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.EventBusClient
}

type processConfigurationDataSourceModel struct {
	processConfigurationBaseModel
}

func (d *processConfigurationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	const resourceName = "EventBus ProcessConfiguration"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId(resourceName),
			"name":        common.SchemaDataSourceName(resourceName),
			"description": common.SchemaDataSourceDescription(resourceName),
			"tags":        common.SchemaDataSourceTags(resourceName),
			"icon_id":     common.SchemaDataSourceIconID(resourceName),

			"destination": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The destination of the %s.", resourceName),
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(func(v string) error {
						return v1.ProcessConfigurationSettingsDestination(v).Validate()
					}),
				},
			},
			"parameters": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The parameter of the %s.", resourceName),
			},
		},
	}
}

func (d *processConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data processConfigurationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	tags := common.TsetToStrings(data.Tags)
	if name == "" && len(tags) == 0 {
		resp.Diagnostics.AddError("Invalid Attribute", "Either name or tags must be specified.")
		return
	}

	processConfigurationOp := eventbus.NewProcessConfigurationOp(d.client)
	pcs, err := processConfigurationOp.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("could not find SakuraCloud EventBus ProcessConfiguration resources: %s", err))
		return
	}

	for _, pc := range pcs {
		if name != "" && pc.Name != name {
			continue
		}

		tagsMatched := true
		for _, tagToFind := range tags {
			if slices.Contains(pc.Tags, tagToFind) {
				continue
			}
			tagsMatched = false
			break
		}
		if !tagsMatched {
			continue
		}

		if err := data.updateState(&pc); err != nil {
			resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to update EventBus ProcessConfiguration[%s] state: %s", data.ID.String(), err))
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError("API Error", fmt.Sprintf("could not find any SakuraCloud EventBus ProcessConfiguration resources with name=%q and tags=%v", name, tags))
}
