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

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/sacloud/eventbus-api-go"
	eventbus_api "github.com/sacloud/eventbus-api-go/apis/v1"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

type eventBusProcessConfigurationResource struct {
	client *eventbus_api.Client
}

var (
	_ resource.Resource                = &eventBusProcessConfigurationResource{}
	_ resource.ResourceWithConfigure   = &eventBusProcessConfigurationResource{}
	_ resource.ResourceWithImportState = &eventBusProcessConfigurationResource{}
)

func eventBusProcessConfigurationResource() resource.Resource {
	return &eventBusProcessConfigurationResource{}
}

func (r *eventBusProcessConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_bus_process_configuration"
}

func (r *eventBusProcessConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.EventBusClient
}

type eventBusProcessConfigurationResourceModel struct {
	eventBusProcessConfigurationBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *eventBusProcessConfigurationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Required:    true,
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

			// TODO: credentialsどうしようかなー
			// TODO: AccessToken, AccessTokenSecret for simplenotification
			// TODO: APIKey for simplemq

			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *eventBusProcessConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *eventBusProcessConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan eventBusProcessConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	processConfigurationOp := eventbus.NewProcessConfigurationOp(r.client)
	// TODO:
	mq, err := processConfigurationOp.Create(ctx, expandEventBusProcessConfigurationCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("create EventBus ProcessConfiguration failed: %s", err))
		return
	}
	pcID := eventbus_api.GetProcessConfiguration(mq)

	// SDK v2ではUpdateを呼び出して更新していたが、Frameworkではアクション間での状態の共有が難しいためメソッドに括り出して処理を共通化
	err = r.callUpdateRequest(ctx, qid, &plan, mq)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	q := getMessageQueue(ctx, r.client, qid, &resp.State, &resp.Diagnostics)
	if q == nil {
		return
	}

	plan.updateState(q)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func expandEventBusProcessConfigurationCreateRequest(d *eventBusProcessConfigurationResourceModel) eventbus_api.ProcessConfigurationRequestSettings {
	req := eventbus_api.ProcessConfigurationRequestSettings{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Settings: eventbus_api.DestinationSettings{
			Destination: eventbus_api.CreateProcessConfigurationRequestDestination(d.Destination.ValueString()),
			Parameters:  d.Parameters.ValueString(),
		},
		Provider: eventbus_api.ProcessConfigurationProvider{
			Class: "eventbusprocessconfiguration",
		},
		// TODO: Icon, Tagsはsdkが対応していないので保留中
	}

	return req
}
