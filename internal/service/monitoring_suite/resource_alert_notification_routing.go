// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	v1 "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
)

type alertNotificationRoutingResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &alertNotificationRoutingResource{}
	_ resource.ResourceWithConfigure   = &alertNotificationRoutingResource{}
	_ resource.ResourceWithImportState = &alertNotificationRoutingResource{}
)

func NewAlertNotificationRoutingResource() resource.Resource {
	return &alertNotificationRoutingResource{}
}

func (r *alertNotificationRoutingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert_notification_routing"
}

func (r *alertNotificationRoutingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type alertNotificationRoutingResourceModel struct {
	alertNotificationRoutingBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *alertNotificationRoutingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":       common.SchemaResourceId("Monitoring Suite Alert Notification Routing"),
			"alert_id": schemaResourceAlertId(),
			"notification_target_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Alert Notification Target.",
			},
			"resend_interval_minutes": schema.Int32Attribute{
				Optional:    true,
				Description: "The resend interval in minutes of the Alert Notification Routing.",
			},
			"match_labels": schema.ListNestedAttribute{
				Required:    true,
				Description: "The list of match label of the Alert Notification Routing.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the match label.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
						"value": schema.StringAttribute{
							Required:    true,
							Description: "The value of the match label.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 256),
							},
						},
					},
				},
			},
			"order": schema.Int32Attribute{
				Computed:    true,
				Description: "The order of the Alert Notification Routing.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: `Manages a Monitoring Suite Alert Notification Routing.`,
	}
}

func (r *alertNotificationRoutingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *alertNotificationRoutingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertNotificationRoutingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewNotificationRoutingOp(r.client)
	created, err := op.Create(ctx, plan.AlertID.ValueString(), monitoringsuite.NotificationRoutingCreateParams{
		NotificationTargetUID: uuid.MustParse(plan.NotificationTargetID.ValueString()),
		ResendIntervalMinutes: expandResendIntervalMinutes(&plan),
		MatchLabels:           expandAlertNotificationRoutingMatchLabels(&plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Alert Notification Routing: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertNotificationRoutingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertNotificationRoutingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	routing := getAlertNotificationRouting(ctx, r.client, state.AlertID.ValueString(), state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if routing == nil {
		return
	}

	state.updateState(routing)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *alertNotificationRoutingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertNotificationRoutingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewNotificationRoutingOp(r.client)
	uid := uuid.MustParse(plan.NotificationTargetID.ValueString())
	updated, err := op.Update(ctx, plan.AlertID.ValueString(), uuid.MustParse(plan.ID.ValueString()), monitoringsuite.NotificationRoutingUpdateParams{
		NotificationTargetUID: &uid,
		ResendIntervalMinutes: expandResendIntervalMinutes(&plan),
		MatchLabels:           expandAlertNotificationRoutingMatchLabels(&plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Alert Notification Target[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertNotificationRoutingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertNotificationRoutingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewNotificationRoutingOp(r.client)
	if err := op.Delete(ctx, state.AlertID.ValueString(), uuid.MustParse(state.ID.ValueString())); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Alert Notification Routing[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getAlertNotificationRouting(ctx context.Context, client *monitoringsuiteapi.Client, projectID, id string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.NotificationRouting {
	op := monitoringsuite.NewNotificationRoutingOp(client)
	routing, err := op.Read(ctx, projectID, uuid.MustParse(id))
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Alert Notification Routing[%s]: %s", id, err))
		return nil
	}
	return routing
}

func expandAlertNotificationRoutingMatchLabels(model *alertNotificationRoutingResourceModel) []v1.MatchLabelsItem {
	var labels []v1.MatchLabelsItem
	for _, l := range model.MatchLabels {
		labels = append(labels, v1.MatchLabelsItem{
			Name:  l.Name.ValueString(),
			Value: l.Value.ValueString(),
		})
	}
	return labels
}

func expandResendIntervalMinutes(model *alertNotificationRoutingResourceModel) *int {
	if !utils.IsKnown(model.ResendIntervalMinutes) {
		return nil
	}

	min := int(model.ResendIntervalMinutes.ValueInt32())
	return &min
}
