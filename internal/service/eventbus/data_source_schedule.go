// Copyright 2016-2025 terraform-provider-sakura authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eventbus

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/eventbus-api-go"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type scheduleDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &scheduleDataSource{}
	_ datasource.DataSourceWithConfigure = &scheduleDataSource{}
)

func NewEventBusScheduleDataSource() datasource.DataSource {
	return &scheduleDataSource{}
}

func (d *scheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eventbus_schedule"
}

func (d *scheduleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.EventBusClient
}

type scheduleDataSourceModel struct {
	scheduleBaseModel
}

func (d *scheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	const resourceName = "EventBus Schedule"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId(resourceName),
			"name":        common.SchemaDataSourceName(resourceName),
			"description": common.SchemaDataSourceDescription(resourceName),
			// TODO: iconはsdkが対応していないので保留中
			// "icon_id":     common.SchemaDataSourceIconID(resourceName),
			// NOTE: tagsは設定がsdk非対応のためTFからは設定不可能だが、UI上などで設定した値はsdkが返してくれているので取得可能。
			"tags": common.SchemaDataSourceTags(resourceName),

			"process_configuration_id": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The ProcessConfiguration ID of the %s.", resourceName),
			},
			"recurring_step": schema.Int64Attribute{
				Computed:    true,
				Description: desc.Sprintf("The RecurringStep of the %s.", resourceName),
			},
			"recurring_unit": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The RecurringUnit of the %s.", resourceName),
			},
			"starts_at": schema.Int64Attribute{
				Computed:    true,
				Description: desc.Sprintf("The start time of the %s. (in epoch milliseconds)", resourceName),
			},
		},
	}
}

func (d *scheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data scheduleDataSourceModel
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

	scheduleOp := eventbus.NewScheduleOp(d.client)
	schedules, err := scheduleOp.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("could not find SakuraCloud EventBus Schedule resource: %s", err))
		return
	}

	for _, s := range schedules {
		if name != "" && s.Name != name {
			continue
		}

		tagsMatched := true
		for _, tagToFind := range tags {
			if slices.Contains(s.Tags, tagToFind) {
				continue
			}
			tagsMatched = false
			break
		}
		if !tagsMatched {
			continue
		}

		data.updateState(&s)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	resp.Diagnostics.AddError("API Error", fmt.Sprintf("could not find any SakuraCloud EventBus Schedule resources with  name=%q and tags=%v", name, tags))
}
