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
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/eventbus-api-go"
	eventbus_api "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakuracloud/internal/validator"
)

type eventBusProcessConfigurationResource struct {
	client *eventbus_api.Client
}

var (
	_ resource.Resource                = &eventBusProcessConfigurationResource{}
	_ resource.ResourceWithConfigure   = &eventBusProcessConfigurationResource{}
	_ resource.ResourceWithImportState = &eventBusProcessConfigurationResource{}
)

func NewEventBusProcessConfigurationResource() resource.Resource {
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

			// TODO: credentialsを見て動的にdestinationをcomputeできると良い？でないとユーザ的には二度手間
			"simplemq_api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: desc.Sprintf("The SimpleMQ API key for %s.", resourceName),
			},
			"simplenotification_api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: desc.Sprintf("The SimpleNotification API key for %s.", resourceName),
			},
			"simplenotification_api_key_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: desc.Sprintf("The SimpleNotification API key secret for %s.", resourceName),
			},

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
	pc, err := processConfigurationOp.Create(ctx, expandEventBusProcessConfigurationCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("create EventBus ProcessConfiguration failed: %s", err))
		return
	}
	pcID := strconv.FormatInt(pc.ID, 10)

	// SDK v2ではUpdateを呼び出して更新していたが、Frameworkではアクション間での状態の共有が難しいためメソッドに括り出して処理を共通化
	err = r.callUpdateSecretRequest(ctx, pcID, &plan, pc)
	if err != nil {
		resp.Diagnostics.AddError("UpdateSecret Error", err.Error())
		return
	}

	gotPC := getProcessConfiguration(ctx, r.client, pcID, &resp.State, &resp.Diagnostics)
	if gotPC == nil {
		return
	}

	plan.updateState(gotPC)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eventBusProcessConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TODO: impl
}

func (r *eventBusProcessConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: impl
}

func (r *eventBusProcessConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TODO: impl
}

func (r *eventBusProcessConfigurationResource) callUpdateSecretRequest(ctx context.Context, id string, plan *eventBusProcessConfigurationResourceModel, pc *eventbus_api.ProcessConfiguration) error {
	var err error
	processConfigurationOp := eventbus.NewProcessConfigurationOp(r.client)

	if pc == nil {
		_, err = processConfigurationOp.Read(ctx, id)
		if err != nil {
			return fmt.Errorf("could not read EventBus ProcessConfiguration[%s]: %w", id, err)
		}
	}

	err = processConfigurationOp.UpdateSecret(ctx, id, expandEventBusUpdateRequest(plan))
	if err != nil {
		return fmt.Errorf("update secret on EventBus ProcessConfiguration[%s] failed: %w", id, err)
	}

	return nil
}

func getProcessConfiguration(ctx context.Context, client *eventbus_api.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *eventbus_api.ProcessConfiguration {
	processConfigurationOp := eventbus.NewProcessConfigurationOp(client)
	pc, err := processConfigurationOp.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Get ProcessConfiguration Error", fmt.Sprintf("could not read EventBus ProcessConfiguration[%s]: %s", id, err))
		return nil
	}

	return pc
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

func expandEventBusUpdateRequest(d *eventBusProcessConfigurationResourceModel) eventbus_api.ProcessConfigurationSecret {
	req := eventbus_api.ProcessConfigurationSecret{}

	if !d.SimpleNotificationAccessToken.IsNull() && !d.SimpleNotificationAccessToken.IsUnknown() {
		req.AccessToken = d.SimpleNotificationAccessToken.ValueString()
	}
	if !d.SimpleNotificationAccessTokenSecret.IsNull() && !d.SimpleNotificationAccessTokenSecret.IsUnknown() {
		req.AccessTokenSecret = d.SimpleNotificationAccessTokenSecret.ValueString()
	}
	if !d.SimpleMQAPIKey.IsNull() && !d.SimpleMQAPIKey.IsUnknown() {
		req.APIKey = d.SimpleMQAPIKey.ValueString()
	}

	return req
}
