// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_notification

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	simplenotification "github.com/sacloud/simple-notification-api-go"
	v1 "github.com/sacloud/simple-notification-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type DestinationDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &DestinationDataSource{}
	_ datasource.DataSourceWithConfigure = &DestinationDataSource{}
)

func NewDestinationDataSource() datasource.DataSource {
	return &DestinationDataSource{}
}

func (d *DestinationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_simple_notification_destination"
}

func (d *DestinationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.SimpleNotificationClient
}

type DestinationDataSourceModel struct {
	destinationBaseModel
}

func (d *DestinationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	const resourceName = "SimpleNotification Destination"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId(resourceName),
			"name":        common.SchemaDataSourceName(resourceName),
			"description": common.SchemaDataSourceDescription(resourceName),
			"tags":        common.SchemaDataSourceTags(resourceName),
			"icon_id":     common.SchemaDataSourceIconID(resourceName),
			"type": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The type of the %s.", resourceName),
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(func(v string) error {
						if err := v1.CommonServiceItemDestinationSettingsType(v).Validate(); err != nil {
							return fmt.Errorf("invalid operator: %s", v)
						}
						return nil
					}),
				},
			},
			"value": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The value of the %s.", resourceName),
			},
		},
		MarkdownDescription: "Get information about an existing SimpleNotification Destination.",
	}
}

func (d *DestinationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DestinationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	tags := common.TsetToStrings(data.Tags)
	if name == "" && len(tags) == 0 {
		resp.Diagnostics.AddError("Read: Attribute Error", "either 'name' or 'tags' must be specified.")
		return
	}

	DestinationOp := simplenotification.NewDestinationOp(d.client)
	destListRes, err := DestinationOp.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to list SimpleNotification ProcessConfiguration resources: %s", err))
		return
	}

	for _, dest := range destListRes.CommonServiceItems {
		if name != "" && dest.Name != name {
			continue
		}

		tagsMatched := true
		for _, tagToFind := range tags {
			if slices.Contains(dest.Tags, tagToFind) {
				continue
			}
			tagsMatched = false
			break
		}
		if !tagsMatched {
			continue
		}

		if err := data.updateState(&dest); err != nil {
			resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update SimpleNotification Destination[%s] state: %s", data.ID.String(), err))
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError("Read: Search Error", fmt.Sprintf("failed to find any SimpleNotification Destination resources with name=%q and tags=%v", name, tags))
}
