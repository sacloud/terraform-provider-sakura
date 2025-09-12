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
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakuracloud/internal/validator"
)

type processConfigurationResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &processConfigurationResource{}
	_ resource.ResourceWithConfigure   = &processConfigurationResource{}
	_ resource.ResourceWithImportState = &processConfigurationResource{}
)

func NewEventBusProcessConfigurationResource() resource.Resource {
	return &processConfigurationResource{}
}

func (r *processConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_bus_process_configuration"
}

func (r *processConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.EventBusClient
}

type processConfigurationResourceModel struct {
	processConfigurationBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *processConfigurationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	const resourceName = "EventBus ProcessConfiguration"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId(resourceName),
			"name":        common.SchemaResourceName(resourceName),
			"description": common.SchemaResourceDescription(resourceName),
			// TODO: icon, tagsはsdkが対応していないので保留中
			"tags": common.SchemaResourceTags(resourceName), // tfsdk tagでエラーになるので定義だけする
			// "icon_id":     common.SchemaResourceIconID(resourceName),

			"destination": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The destination of the %s.", resourceName),
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(func(v string) error {
						return v1.ProcessConfigurationDestination(v).Validate()
					}),
				},
			},
			"parameters": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The parameter of the %s.", resourceName),
			},

			// TODO: some extra fields
			// group_id, message
			// queue_name, content
			// ref: https://manual.sakura.ad.jp/cloud/appliance/eventbus/index.html#id16

			// TODO: credentialsを見て動的にdestinationをcomputeできると良い？でないとユーザ的には二度手間
			"simplemq_api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: desc.Sprintf("The SimpleMQ API key for %s.", resourceName),
			},
			"simplenotification_access_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: desc.Sprintf("The SimpleNotification access token for %s.", resourceName),
			},
			"simplenotification_access_token_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: desc.Sprintf("The SimpleNotification access token secret for %s.", resourceName),
			},

			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *processConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *processConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan processConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	processConfigurationOp := eventbus.NewProcessConfigurationOp(r.client)
	pc, err := processConfigurationOp.Create(ctx, expandProcessConfigurationCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("create EventBus ProcessConfiguration failed: %s", err))
		return
	}
	pcID := strconv.FormatInt(pc.ID, 10)

	// SDK v2ではUpdateを呼び出して更新していたが、Frameworkではアクション間での状態の共有が難しいためメソッドに括り出して処理を共通化
	err = r.callProcessConfigurationUpdateSecretRequest(ctx, pcID, &plan, pc)
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

func (r *processConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state processConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pc := getProcessConfiguration(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if pc == nil {
		return
	}

	state.updateState(pc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *processConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan processConfigurationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	processConfigurationOp := eventbus.NewProcessConfigurationOp(r.client)

	// TODO: expandProcessConfigurationCreateRequestでいいか確認。たまたま全部Requiredだから良いかも知れないが、一応。
	if _, err := processConfigurationOp.Update(ctx, plan.ID.ValueString(), expandProcessConfigurationCreateRequest(&plan)); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("update on EventBus ProcessConfiguration[%s] failed: %s", plan.ID.ValueString(), err))
		return
	}

	if err := r.callProcessConfigurationUpdateSecretRequest(ctx, plan.ID.ValueString(), &plan, nil); err != nil {
		resp.Diagnostics.AddError("UpdateSecret Error", err.Error())
		return
	}

	pc := getProcessConfiguration(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if pc == nil {
		return
	}

	plan.updateState(pc)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *processConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state processConfigurationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	processConfigurationOp := eventbus.NewProcessConfigurationOp(r.client)
	pc := getProcessConfiguration(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if pc == nil {
		return
	}

	if err := processConfigurationOp.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("delete EventBus ProcessConfiguration[%s] failed: %s", state.ID.ValueString(), err))
		return
	}
}

func (r *processConfigurationResource) callProcessConfigurationUpdateSecretRequest(ctx context.Context, id string, plan *processConfigurationResourceModel, pc *v1.ProcessConfiguration) error {
	var err error
	processConfigurationOp := eventbus.NewProcessConfigurationOp(r.client)

	if pc == nil {
		_, err = processConfigurationOp.Read(ctx, id)
		if err != nil {
			return fmt.Errorf("could not read EventBus ProcessConfiguration[%s]: %w", id, err)
		}
	}

	err = processConfigurationOp.UpdateSecret(ctx, id, expandProcessConfigurationUpdateSecretRequest(plan))
	if err != nil {
		return fmt.Errorf("update secret on EventBus ProcessConfiguration[%s] failed: %w", id, err)
	}

	return nil
}

func getProcessConfiguration(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.ProcessConfiguration {
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

func expandProcessConfigurationCreateRequest(d *processConfigurationResourceModel) v1.ProcessConfigurationRequestSettings {
	req := v1.ProcessConfigurationRequestSettings{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Settings: v1.DestinationSettings{
			Destination: v1.CreateProcessConfigurationRequestDestination(d.Destination.ValueString()),
			Parameters:  d.Parameters.ValueString(),
		},
		Provider: v1.ProcessConfigurationProvider{
			Class: "eventbusprocessconfiguration",
		},
		// TODO: Icon, Tagsはsdkが対応していないので保留中
	}

	return req
}

func expandProcessConfigurationUpdateSecretRequest(d *processConfigurationResourceModel) v1.ProcessConfigurationSecret {
	req := v1.ProcessConfigurationSecret{}

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

type scheduleResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &scheduleResource{}
	_ resource.ResourceWithConfigure   = &scheduleResource{}
	_ resource.ResourceWithImportState = &scheduleResource{}
)

func NewEventBusScheduleResource() resource.Resource {
	return &scheduleResource{}
}

func (r *scheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_bus_schedule"
}

func (r *scheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.EventBusClient
}

type scheduleResourceModel struct {
	scheduleBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *scheduleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	const resourceName = "EventBus Schedule"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId(resourceName),
			"name":        common.SchemaResourceName(resourceName),
			"description": common.SchemaResourceDescription(resourceName),
			// TODO: icon, tagsはsdkが対応していないので保留中
			// "tags":        common.SchemaResourceTags(resourceName),
			// "icon_id":     common.SchemaResourceIconID(resourceName),

			"process_configuration_id": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The ProcessConfiguration ID of the %s.", resourceName),
			},
			"recurring_step": schema.Int64Attribute{
				Optional:    true,
				Description: desc.Sprintf("The RecurringStep of the %s.", resourceName),
			},
			"recurring_unit": schema.StringAttribute{
				Optional:    true,
				Description: desc.Sprintf("The RecurringUnit of the %s.", resourceName),
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(func(v string) error {
						return v1.ScheduleRecurringUnit(v).Validate()
					}),
				},
			},
			"starts_at": schema.Int64Attribute{
				Required:    true,
				Description: desc.Sprintf("The start time of the %s. (in epoch milliseconds)", resourceName),
			},

			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *scheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *scheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	scheduleOp := eventbus.NewScheduleOp(r.client)
	schedule, err := scheduleOp.Create(ctx, expandScheduleCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("create EventBus Schedule failed: %s", err))
		return
	}
	scheduleID := strconv.FormatInt(schedule.ID, 10)

	gotSchedule := getSchedule(ctx, r.client, scheduleID, &resp.State, &resp.Diagnostics)
	if gotSchedule == nil {
		return
	}

	plan.updateState(gotSchedule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedule := getSchedule(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if schedule == nil {
		return
	}

	state.updateState(schedule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan scheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	scheduleOp := eventbus.NewScheduleOp(r.client)

	// TODO: expandScheduleCreateRequestでいいか確認。たまたま全部Requiredだから良いかも知れないが、一応。
	if _, err := scheduleOp.Update(ctx, plan.ID.ValueString(), expandScheduleCreateRequest(&plan)); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("update on EventBus Schedule[%s] failed: %s", plan.ID.ValueString(), err))
		return
	}

	schedule := getSchedule(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if schedule == nil {
		return
	}

	plan.updateState(schedule)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scheduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	scheduleOp := eventbus.NewScheduleOp(r.client)
	schedule := getSchedule(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if schedule == nil {
		return
	}

	if err := scheduleOp.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("delete EventBus Schedule[%s] failed: %s", state.ID.ValueString(), err))
		return
	}
}

func getSchedule(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.Schedule {
	scheduleOp := eventbus.NewScheduleOp(client)
	schedule, err := scheduleOp.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Get Schedule Error", fmt.Sprintf("could not read EventBus Schedule[%s]: %s", id, err))
		return nil
	}

	return schedule
}

func expandScheduleCreateRequest(d *scheduleResourceModel) v1.ScheduleRequestSettings {
	req := v1.ScheduleRequestSettings{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		// TODO: Icon, Tagsはsdkが対応していないので保留中

		Settings: v1.ScheduleSettings{
			ProcessConfigurationID: d.ProcessConfigurationID.ValueString(),
			RecurringStep:          int(d.RecurringStep.ValueInt64()),
			RecurringUnit:          v1.CreateScheduleRequestRecurringUnit(d.RecurringUnit.ValueString()),
			StartsAt:               d.StartsAt.ValueInt64(),
		},
		Provider: v1.ScheduleProvider{
			Class: "eventbusschedule",
		},
	}

	return req
}
