// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package auto_scale

import (
	"context"
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	autoScaler "github.com/sacloud/autoscaler/core"
	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type autoScaleResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &autoScaleResource{}
	_ resource.ResourceWithConfigure   = &autoScaleResource{}
	_ resource.ResourceWithImportState = &autoScaleResource{}
)

func NewAutoScaleResource() resource.Resource {
	return &autoScaleResource{}
}

func (r *autoScaleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auto_scale"
}

func (r *autoScaleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type autoScaleResourceModel struct {
	autoScaleBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *autoScaleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("AutoScale"),
			"name":        common.SchemaResourceName("AutoScale"),
			"description": common.SchemaResourceDescription("AutoScale"),
			"tags":        common.SchemaResourceTags("AutoScale"),
			"icon_id":     common.SchemaResourceIconID("AutoScale"),
			"zones": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "Set of zone names where monitored resources are located",
			},
			"config": schema.StringAttribute{
				Required:    true,
				Description: "The configuration file for sacloud/autoscaler",
				Validators: []validator.String{
					sacloudvalidator.StringFuncValidator(func(v string) error {
						config := autoScaler.Config{}
						err := yaml.UnmarshalWithOptions([]byte(v), &config, yaml.Strict())
						if err != nil {
							return errors.New(yaml.FormatError(err, false, true))
						}
						return nil
					}),
				},
			},
			"api_key_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the API key",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether to enable AutoScale",
			},
			"trigger_type": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("Trigger type of schedule. This must be one of [%s]", []string{"cpu", "router", "schedule", "none"}),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"cpu", "router", "schedule", "none"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"router_threshold_scaling": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"router_prefix": schema.StringAttribute{
						Required:    true,
						Description: "Router name prefix to be monitored",
					},
					"direction": schema.StringAttribute{
						Required:    true,
						Description: desc.Sprintf("This must be one of [%s]", []string{"in", "out"}),
						Validators: []validator.String{
							stringvalidator.OneOf([]string{"in", "out"}...),
						},
					},
					"mbps": schema.Int32Attribute{
						Required:    true,
						Description: "Mbps",
					},
				},
			},
			"cpu_threshold_scaling": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"server_prefix": schema.StringAttribute{
						Required:    true,
						Description: "Server name prefix to be monitored",
					},
					"up": schema.Int32Attribute{
						Required:    true,
						Description: "Threshold for average CPU utilization to scale up/out",
					},
					"down": schema.Int32Attribute{
						Required:    true,
						Description: "Threshold for average CPU utilization to scale down/in",
					},
				},
			},
			"schedule_scaling": schema.ListNestedAttribute{
				Optional: true,
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 2),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Required:    true,
							Description: desc.Sprintf("This must be one of [%s]", []string{"up", "down"}),
							Validators: []validator.String{
								stringvalidator.OneOf([]string{"up", "down"}...),
							},
						},
						"hour": schema.Int32Attribute{
							Required:    true,
							Description: "Hour to be triggered",
							Validators: []validator.Int32{
								int32validator.Between(0, 23),
							},
						},
						"minute": schema.Int32Attribute{
							Required:    true,
							Description: desc.Sprintf("Minute to be triggered. This must be one of [%s]", []string{"0", "15", "30", "45"}),
							Validators: []validator.Int32{
								int32validator.OneOf(0, 15, 30, 45),
							},
						},
						"days_of_week": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: desc.Sprintf("A set of days of week to backed up. The values in the list must be in [%s]", iaastypes.DaysOfTheWeekStrings),
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(stringvalidator.OneOf(iaastypes.DaysOfTheWeekStrings...)),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a AutoScale.",
	}
}

func (r *autoScaleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *autoScaleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan autoScaleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	autoScaleOp := iaas.NewAutoScaleOp(r.client)
	as, err := autoScaleOp.Create(ctx, expandAutoScaleCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create AutoScale: %s", err))
		return
	}

	plan.updateState(as)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *autoScaleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state autoScaleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	as := getAutoScale(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.updateState(as)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *autoScaleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan autoScaleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	current := getAutoScale(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if current == nil {
		return
	}

	if _, err := iaas.NewAutoScaleOp(r.client).Update(ctx, current.ID, expandAutoScaleUpdateRequest(&plan, current)); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update AutoScale[%s]: %s", current.ID, err))
		return
	}

	as := getAutoScale(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if as == nil {
		return
	}

	plan.updateState(as)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *autoScaleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state autoScaleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // get current state
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	as := getAutoScale(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if as == nil {
		return
	}

	if err := iaas.NewAutoScaleOp(r.client).Delete(ctx, as.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete AutoScale[%s]: %s", as.ID, err))
		return
	}
}

func getAutoScale(ctx context.Context, client *common.APIClient, id string, state *tfsdk.State, diags *diag.Diagnostics) *iaas.AutoScale {
	autoScaleOp := iaas.NewAutoScaleOp(client)
	as, err := autoScaleOp.Read(ctx, common.SakuraCloudID(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read AutoScale[%s]: %s", id, err))
		return nil
	}
	return as
}

func expandAutoScaleCreateRequest(model *autoScaleResourceModel) *iaas.AutoScaleCreateRequest {
	return &iaas.AutoScaleCreateRequest{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Tags:        common.TsetToStrings(model.Tags),
		IconID:      common.ExpandSakuraCloudID(model.IconID),

		Zones:                  common.TsetToStrings(model.Zones),
		Config:                 model.Config.ValueString(),
		APIKeyID:               model.APIKeyID.ValueString(),
		Disabled:               !model.Enabled.ValueBool(),
		TriggerType:            iaastypes.EAutoScaleTriggerType(model.TriggerType.ValueString()),
		CPUThresholdScaling:    expandAutoScaleCPUThresholdScaling(model),
		RouterThresholdScaling: expandAutoScaleRouterThresholdScaling(model),
		ScheduleScaling:        expandAutoScaleScheduleScaling(model),
	}
}

func expandAutoScaleUpdateRequest(model *autoScaleResourceModel, current *iaas.AutoScale) *iaas.AutoScaleUpdateRequest {
	return &iaas.AutoScaleUpdateRequest{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		Tags:        common.TsetToStrings(model.Tags),
		IconID:      common.ExpandSakuraCloudID(model.IconID),

		Zones:                  common.TsetToStrings(model.Zones),
		Config:                 model.Config.ValueString(),
		Disabled:               !model.Enabled.ValueBool(),
		TriggerType:            iaastypes.EAutoScaleTriggerType(model.TriggerType.ValueString()),
		CPUThresholdScaling:    expandAutoScaleCPUThresholdScaling(model),
		RouterThresholdScaling: expandAutoScaleRouterThresholdScaling(model),
		ScheduleScaling:        expandAutoScaleScheduleScaling(model),
		SettingsHash:           current.SettingsHash,
	}
}

func expandAutoScaleCPUThresholdScaling(model *autoScaleResourceModel) *iaas.AutoScaleCPUThresholdScaling {
	if model.CPUThresholdScaling != nil {
		return &iaas.AutoScaleCPUThresholdScaling{
			ServerPrefix: model.CPUThresholdScaling.ServerPrefix.ValueString(),
			Up:           int(model.CPUThresholdScaling.Up.ValueInt32()),
			Down:         int(model.CPUThresholdScaling.Down.ValueInt32()),
		}
	}
	return nil
}

func expandAutoScaleRouterThresholdScaling(model *autoScaleResourceModel) *iaas.AutoScaleRouterThresholdScaling {
	if model.RouterThresholdScaling != nil {
		return &iaas.AutoScaleRouterThresholdScaling{
			RouterPrefix: model.RouterThresholdScaling.RouterPrefix.ValueString(),
			Direction:    model.RouterThresholdScaling.Direction.ValueString(),
			Mbps:         int(model.RouterThresholdScaling.Mbps.ValueInt32()),
		}
	}
	return nil
}

func expandAutoScaleScheduleScaling(model *autoScaleResourceModel) []*iaas.AutoScaleScheduleScaling {
	if model.ScheduleScaling != nil {
		var scheduleScaling []*iaas.AutoScaleScheduleScaling
		for _, ss := range model.ScheduleScaling {
			scheduleScaling = append(scheduleScaling, &iaas.AutoScaleScheduleScaling{
				Action:    iaastypes.EAutoScaleAction(ss.Action.ValueString()),
				Hour:      int(ss.Hour.ValueInt32()),
				Minute:    int(ss.Minute.ValueInt32()),
				DayOfWeek: expandAutoScaleDaysOfWeek(&ss),
			})
		}
		return scheduleScaling
	}
	return nil
}

func expandAutoScaleDaysOfWeek(ss *autoScaleScheduleScalingModel) []iaastypes.EDayOfTheWeek {
	var vs []iaastypes.EDayOfTheWeek
	for _, v := range common.TsetToStrings(ss.DaysOfWeek) {
		vs = append(vs, iaastypes.EDayOfTheWeek(v))
	}
	iaastypes.SortDayOfTheWeekList(vs)
	return vs
}
