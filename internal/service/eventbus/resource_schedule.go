// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/eventbus-api-go"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

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
	resp.TypeName = req.ProviderTypeName + "_eventbus_schedule"
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
			"tags":        common.SchemaResourceTags(resourceName),
			"icon_id":     common.SchemaResourceIconID(resourceName),

			"process_configuration_id": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The ProcessConfiguration ID of the %s.", resourceName),
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
			},
			"recurring_step": schema.Int64Attribute{
				Optional:    true,
				Description: desc.Sprintf("The RecurringStep of the %s.", resourceName),
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("recurring_unit")),
					int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("crontab")),
				},
			},
			"recurring_unit": schema.StringAttribute{
				Optional:    true,
				Description: desc.Sprintf("The RecurringUnit of the %s.", resourceName),
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("recurring_step")),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("crontab")),
					sacloudvalidator.StringFuncValidator(func(v string) error {
						return v1.ScheduleSettingsRecurringUnit(v).Validate()
					}),
				},
			},
			"crontab": schema.StringAttribute{
				Optional:    true,
				Description: desc.Sprintf("Crontab of the %s.", resourceName),
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName("recurring_step"),
						path.MatchRelative().AtParent().AtName("recurring_unit"),
					),
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
		MarkdownDescription: "Manages a EventBus Schedule.",
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
	scheduleID := schedule.ID

	gotSchedule := getSchedule(ctx, r.client, scheduleID, &resp.State, &resp.Diagnostics)
	if gotSchedule == nil {
		return
	}

	if err := plan.updateState(gotSchedule); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to update EventBus Schedule[%s] state: %s", plan.ID.String(), err))
		return
	}
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

	if err := state.updateState(schedule); err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to update EventBus Schedule[%s] state: %s", state.ID.String(), err))
		return
	}
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

	if _, err := scheduleOp.Update(ctx, plan.ID.ValueString(), expandScheduleUpdateRequest(&plan)); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("update on EventBus Schedule[%s] failed: %s", plan.ID.ValueString(), err))
		return
	}

	schedule := getSchedule(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if schedule == nil {
		return
	}

	if err := plan.updateState(schedule); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update EventBus Schedule[%s] state: %s", plan.ID.String(), err))
		return
	}
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

func getSchedule(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.CommonServiceItem {
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

func expandScheduleCreateRequest(d *scheduleResourceModel) v1.CreateCommonServiceItemRequest {
	req := v1.CreateCommonServiceItemRequest{
		CommonServiceItem: v1.CreateCommonServiceItemRequestCommonServiceItem{
			Name:        d.Name.ValueString(),
			Description: v1.NewOptNilString(d.Description.ValueString()),
			Settings: v1.NewScheduleSettingsSettings(v1.ScheduleSettings{
				ProcessConfigurationID: d.ProcessConfigurationID.ValueString(),
				StartsAt:               v1.NewInt64ScheduleSettingsStartsAt(d.StartsAt.ValueInt64()),
			}),
			Provider: v1.Provider{
				Class: v1.ProviderClassEventbusschedule,
			},
			Tags: common.TsetToStrings(d.Tags),
		},
	}

	if !d.RecurringStep.IsNull() && !d.RecurringStep.IsUnknown() {
		req.CommonServiceItem.Settings.ScheduleSettings.RecurringStep = v1.NewOptInt(int(d.RecurringStep.ValueInt64()))
	}
	if ru := d.RecurringUnit.ValueString(); ru != "" {
		req.CommonServiceItem.Settings.ScheduleSettings.RecurringUnit = v1.NewOptScheduleSettingsRecurringUnit(v1.ScheduleSettingsRecurringUnit(ru))
	}

	if crontab := d.Crontab.ValueString(); crontab != "" {
		req.CommonServiceItem.Settings.ScheduleSettings.Crontab = v1.NewOptString(crontab)
	}

	if !d.IconID.IsNull() && !d.IconID.IsUnknown() {
		req.CommonServiceItem.Icon = v1.NewOptNilIcon(v1.Icon{
			ID: v1.NewOptString(d.IconID.ValueString()),
		})
	}

	return req
}

func expandScheduleUpdateRequest(d *scheduleResourceModel) v1.UpdateCommonServiceItemRequest {
	req := v1.UpdateCommonServiceItemRequest{
		CommonServiceItem: v1.UpdateCommonServiceItemRequestCommonServiceItem{
			Name:        v1.NewOptString(d.Name.ValueString()),
			Description: v1.NewOptNilString(d.Description.ValueString()),
			Settings: v1.NewOptSettings(
				v1.NewScheduleSettingsSettings(v1.ScheduleSettings{
					ProcessConfigurationID: d.ProcessConfigurationID.ValueString(),
					StartsAt:               v1.NewInt64ScheduleSettingsStartsAt(d.StartsAt.ValueInt64()),
				}),
			),
			Provider: v1.NewOptProvider(
				v1.Provider{
					Class: v1.ProviderClassEventbusprocessconfiguration,
				},
			),
			Tags: common.TsetToStrings(d.Tags),
		},
	}

	if !d.RecurringStep.IsNull() && !d.RecurringStep.IsUnknown() {
		req.CommonServiceItem.Settings.Value.ScheduleSettings.RecurringStep = v1.NewOptInt(int(d.RecurringStep.ValueInt64()))
	}
	if ru := d.RecurringUnit.ValueString(); ru != "" {
		req.CommonServiceItem.Settings.Value.ScheduleSettings.RecurringUnit = v1.NewOptScheduleSettingsRecurringUnit(v1.ScheduleSettingsRecurringUnit(ru))
	}

	if crontab := d.Crontab.ValueString(); crontab != "" {
		req.CommonServiceItem.Settings.Value.ScheduleSettings.Crontab = v1.NewOptString(crontab)
	}

	if !d.IconID.IsNull() && !d.IconID.IsUnknown() {
		req.CommonServiceItem.Icon = v1.NewOptNilIcon(v1.Icon{
			ID: v1.NewOptString(d.IconID.ValueString()),
		})
	}

	return req
}
