// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/eventbus-api-go"
	v1 "github.com/sacloud/eventbus-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type triggerResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &triggerResource{}
	_ resource.ResourceWithConfigure   = &triggerResource{}
	_ resource.ResourceWithImportState = &triggerResource{}
)

func NewEventBusTriggerResource() resource.Resource {
	return &triggerResource{}
}

func (r *triggerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eventbus_trigger"
}

func (r *triggerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.EventBusClient
}

type triggerResourceModel struct {
	triggerBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *triggerResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	const resourceName = "EventBus Trigger"
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
			"source": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The source of the %s.", resourceName),
			},
			"types": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: desc.Sprintf("The types of the %s.", resourceName),
			},
			"conditions": schema.ListNestedAttribute{
				Optional:    true,
				Description: desc.Sprintf("The conditions of the %s.", resourceName),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: desc.Sprintf("The key of the condition for %s.", resourceName),
						},
						"op": schema.StringAttribute{
							Required:    true,
							Description: desc.Sprintf("The operator of the condition for %s.", resourceName),
							Validators: []validator.String{
								sacloudvalidator.StringFuncValidator(func(v string) error {
									if t := v1.TriggerSettingsConditionsItemType(v); t == v1.TriggerConditionEqTriggerSettingsConditionsItem || t == v1.TriggerConditionInTriggerSettingsConditionsItem {
										return nil
									}
									return fmt.Errorf("invalid operator: %s", v)
								}),
							},
						},
						"values": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: desc.Sprintf("The values of the condition for %s. Length shoud be 1 when `op` is `eq`, and at least 1 when `op` is `in`.", resourceName),
						},
					},
					Validators: []validator.Object{
						sacloudvalidator.ObjectFuncValidator(func(o types.Object) error {
							key := o.Attributes()["key"].(types.String).ValueString()
							op := o.Attributes()["op"].(types.String).ValueString()
							values := o.Attributes()["values"].(types.Set)

							var cond v1.TriggerSettingsConditionsItem
							switch op {
							case string(v1.TriggerConditionEqTriggerSettingsConditionsItem):
								cond = v1.NewTriggerConditionEqTriggerSettingsConditionsItem(v1.TriggerConditionEq{
									Key:    key,
									Op:     v1.TriggerConditionEqOpEq,
									Values: common.TsetToStrings(values),
								})
							case string(v1.TriggerConditionInTriggerSettingsConditionsItem):
								cond = v1.NewTriggerConditionInTriggerSettingsConditionsItem(v1.TriggerConditionIn{
									Key:    key,
									Op:     v1.TriggerConditionInOpIn,
									Values: common.TsetToStrings(values),
								})
							default:
								return errors.New("invalid operator for condition")
							}

							if err := cond.Validate(); err != nil {
								return fmt.Errorf("invalid condition: %w", err)
							}
							return nil
						}),
					},
				},
			},

			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a EventBus Trigger.",
	}
}

func (r *triggerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *triggerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan triggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	triggerOp := eventbus.NewTriggerOp(r.client)
	trigger, err := triggerOp.Create(ctx, expandTriggerCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create EventBus Trigger: %s", err))
		return
	}

	gotTrigger := getTrigger(ctx, r.client, trigger.ID, &resp.State, &resp.Diagnostics)
	if gotTrigger == nil {
		return
	}

	if err := plan.updateState(gotTrigger); err != nil {
		resp.Diagnostics.AddError("Create: Terraform Error", fmt.Sprintf("failed to update EventBus Trigger[%s] state: %s", plan.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *triggerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state triggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	trigger := getTrigger(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if trigger == nil {
		return
	}

	if err := state.updateState(trigger); err != nil {
		resp.Diagnostics.AddError("Read: Terraform Error", fmt.Sprintf("failed to update EventBus Trigger[%s] state: %s", state.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *triggerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan triggerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	triggerOp := eventbus.NewTriggerOp(r.client)

	if _, err := triggerOp.Update(ctx, plan.ID.ValueString(), expandTriggerUpdateRequest(&plan)); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update EventBus Trigger[%s]: %s", plan.ID.String(), err))
		return
	}

	trigger := getTrigger(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if trigger == nil {
		return
	}

	if err := plan.updateState(trigger); err != nil {
		resp.Diagnostics.AddError("Update: Terraform Error", fmt.Sprintf("failed to update EventBus Trigger[%s] state: %s", plan.ID.String(), err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *triggerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state triggerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	triggerOp := eventbus.NewTriggerOp(r.client)
	trigger := getTrigger(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if trigger == nil {
		return
	}

	if err := triggerOp.Delete(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete EventBus Trigger[%s]: %s", state.ID.String(), err))
		return
	}
}

func getTrigger(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.CommonServiceItem {
	triggerOp := eventbus.NewTriggerOp(client)
	trigger, err := triggerOp.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read EventBus Trigger[%s]: %s", id, err))
		return nil
	}

	return trigger
}

func expandTriggerCreateRequest(d *triggerResourceModel) v1.CreateCommonServiceItemRequest {
	req := v1.CreateCommonServiceItemRequest{
		CommonServiceItem: v1.CreateCommonServiceItemRequestCommonServiceItem{
			Name:        d.Name.ValueString(),
			Description: v1.NewOptNilString(d.Description.ValueString()),
			Settings: v1.NewTriggerSettingsSettings(v1.TriggerSettings{
				ProcessConfigurationID: d.ProcessConfigurationID.ValueString(),
				Source:                 d.Source.ValueString(),
			}),
			Provider: v1.Provider{
				Class: v1.ProviderClassEventbustrigger,
			},
			Tags: common.TsetToStrings(d.Tags),
		},
	}

	if !d.Types.IsNull() && !d.Types.IsUnknown() {
		req.CommonServiceItem.Settings.TriggerSettings.Types = v1.NewOptNilStringArray(common.TsetToStrings(d.Types))
	}

	if len(d.Conditions) > 0 {
		conditions := make([]v1.TriggerSettingsConditionsItem, 0, len(d.Conditions))
		for _, c := range d.Conditions {
			switch c.Operator.ValueString() {
			case string(v1.TriggerConditionEqTriggerSettingsConditionsItem):
				conditions = append(conditions, v1.NewTriggerConditionEqTriggerSettingsConditionsItem(v1.TriggerConditionEq{
					Key:    c.Key.ValueString(),
					Op:     v1.TriggerConditionEqOpEq,
					Values: common.TsetToStrings(c.Values),
				}))
			case string(v1.TriggerConditionInTriggerSettingsConditionsItem):
				conditions = append(conditions, v1.NewTriggerConditionInTriggerSettingsConditionsItem(v1.TriggerConditionIn{
					Key:    c.Key.ValueString(),
					Op:     v1.TriggerConditionInOpIn,
					Values: common.TsetToStrings(c.Values),
				}))
			}
		}
		req.CommonServiceItem.Settings.TriggerSettings.Conditions = v1.NewOptNilTriggerSettingsConditionsItemArray(conditions)
	}

	if !d.IconID.IsNull() && !d.IconID.IsUnknown() {
		req.CommonServiceItem.Icon = v1.NewOptNilIcon(v1.Icon{
			ID: v1.NewOptString(d.IconID.ValueString()),
		})
	}

	return req
}

func expandTriggerUpdateRequest(d *triggerResourceModel) v1.UpdateCommonServiceItemRequest {
	req := v1.UpdateCommonServiceItemRequest{
		CommonServiceItem: v1.UpdateCommonServiceItemRequestCommonServiceItem{
			Name:        v1.NewOptString(d.Name.ValueString()),
			Description: v1.NewOptNilString(d.Description.ValueString()),
			Settings: v1.NewOptSettings(
				v1.NewTriggerSettingsSettings(v1.TriggerSettings{
					Source:                 d.Source.ValueString(),
					ProcessConfigurationID: d.ProcessConfigurationID.ValueString(),
				}),
			),
			Provider: v1.NewOptProvider(
				v1.Provider{
					Class: v1.ProviderClassEventbustrigger,
				},
			),
			Tags: common.TsetToStrings(d.Tags),
		},
	}

	if !d.Types.IsNull() && !d.Types.IsUnknown() {
		req.CommonServiceItem.Settings.Value.TriggerSettings.Types = v1.NewOptNilStringArray(common.TsetToStrings(d.Types))
	}

	if len(d.Conditions) > 0 {
		conditions := make([]v1.TriggerSettingsConditionsItem, 0, len(d.Conditions))
		for _, c := range d.Conditions {
			switch c.Operator.ValueString() {
			case string(v1.TriggerConditionEqTriggerSettingsConditionsItem):
				conditions = append(conditions, v1.NewTriggerConditionEqTriggerSettingsConditionsItem(v1.TriggerConditionEq{
					Key:    c.Key.ValueString(),
					Op:     v1.TriggerConditionEqOpEq,
					Values: common.TsetToStrings(c.Values),
				}))
			case string(v1.TriggerConditionInTriggerSettingsConditionsItem):
				conditions = append(conditions, v1.NewTriggerConditionInTriggerSettingsConditionsItem(v1.TriggerConditionIn{
					Key:    c.Key.ValueString(),
					Op:     v1.TriggerConditionInOpIn,
					Values: common.TsetToStrings(c.Values),
				}))
			}
		}
		req.CommonServiceItem.Settings.Value.TriggerSettings.Conditions = v1.NewOptNilTriggerSettingsConditionsItemArray(conditions)
	}

	if !d.IconID.IsNull() && !d.IconID.IsUnknown() {
		req.CommonServiceItem.Icon = v1.NewOptNilIcon(v1.Icon{
			ID: v1.NewOptString(d.IconID.ValueString()),
		})
	}

	return req
}
