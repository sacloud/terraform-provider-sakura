// Copyright 2016-2025 terraform-provider-sakuracloud authors
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

package event_bus

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/eventbus-api-go"
	eventbus_api "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakuracloud/internal/validator"
)

type processConfigurationDataSource struct {
	client *eventbus_api.Client
}

var (
	_ datasource.DataSource              = &processConfigurationDataSource{}
	_ datasource.DataSourceWithConfigure = &processConfigurationDataSource{}
)

func NewEventBusProcessConfigurationDataSource() datasource.DataSource {
	return &processConfigurationDataSource{}
}

func (d *processConfigurationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_bus_process_configuration"
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
			"id":          common.SchemaResourceId(resourceName),
			"name":        common.SchemaResourceName(resourceName),
			"description": common.SchemaResourceDescription(resourceName),
			// TODO: icon, tagsはsdkが対応していないので保留中
			// "tags":        common.SchemaResourceTags(resourceName),
			// "icon_id":     common.SchemaResourceIconID(resourceName),

			"destination": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The destination of the %s.", resourceName),
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(func(v string) error {
						return eventbus_api.ProcessConfigurationDestination(v).Validate()
					}),
				},
			},
			"parameters": schema.StringAttribute{
				Required:    true,
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

	id := data.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Invalid Attribute", "ID must be specified.")
		return
	}

	processConfigurationOp := eventbus.NewProcessConfigurationOp(d.client)
	pc, err := processConfigurationOp.Read(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("could not find SakuraCloud EventBus ProcessConfiguration[%s] resource: %s", id, err))
		return
	}

	data.updateState(pc)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
