// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	api "github.com/sacloud/api-client-go"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	v1 "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type alertNotificationTargetResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &alertNotificationTargetResource{}
	_ resource.ResourceWithConfigure   = &alertNotificationTargetResource{}
	_ resource.ResourceWithImportState = &alertNotificationTargetResource{}
)

func NewAlertNotificationTargetResource() resource.Resource {
	return &alertNotificationTargetResource{}
}

func (r *alertNotificationTargetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_alert_notification_target"
}

func (r *alertNotificationTargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type alertNotificationTargetResourceModel struct {
	alertNotificationTargetBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *alertNotificationTargetResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Monitoring Suite Alert Notification Target"),
			"description": common.SchemaResourceDescription("Monitoring Suite Alert Notification Target"),
			"alert_id":    schemaResourceAlertId(),
			"service_type": schema.StringAttribute{
				Required:    true,
				Description: "The service type of the Alert Notification Target.",
				Validators: []validator.String{
					stringvalidator.OneOf("simple_notification", "eventbus"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Optional:    true,
				Description: "The URL of the Alert Notification Target.",
			},
			"config": schema.StringAttribute{
				Computed:    true,
				Description: "The config of the Alert Notification Target.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: `Manages a Monitoring Suite Alert Notification Target.`,
	}
}

func (r *alertNotificationTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *alertNotificationTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan alertNotificationTargetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewNotificationTargetOp(r.client)
	created, err := op.Create(ctx, plan.AlertID.ValueString(), monitoringsuite.NotificationTargetCreateParams{
		Description: expandOptionalString(plan.Description),
		ServiceType: expandAlertNotificationTargetServiceType(plan.ServiceType.ValueString()),
		URL:         expandNotificationTargetURL(plan.URL.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Alert Project: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertNotificationTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertNotificationTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	target := getAlertNotificationTarget(ctx, r.client, state.AlertID.ValueString(), state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if target == nil {
		return
	}

	state.updateState(target)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *alertNotificationTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertNotificationTargetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewNotificationTargetOp(r.client)
	st := v1.PatchedNotificationTargetRequestServiceType(expandAlertNotificationTargetServiceType(plan.ServiceType.ValueString()))
	updated, err := op.Update(ctx, plan.AlertID.ValueString(), uuid.MustParse(plan.ID.ValueString()), monitoringsuite.NotificationTargetUpdateParams{
		Description: expandOptionalString(plan.Description),
		ServiceType: &st,
		URL:         expandOptionalString(plan.URL),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Alert Notification Target[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *alertNotificationTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertNotificationTargetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewNotificationTargetOp(r.client)
	if err := op.Delete(ctx, state.AlertID.ValueString(), uuid.MustParse(state.ID.ValueString())); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Alert Notification Target[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getAlertNotificationTarget(ctx context.Context, client *monitoringsuiteapi.Client, projectID, id string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.NotificationTarget {
	op := monitoringsuite.NewNotificationTargetOp(client)
	target, err := op.Read(ctx, projectID, uuid.MustParse(id))
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Alert Notification Target[%s]: %s", id, err))
		return nil
	}
	return target
}

func expandAlertNotificationTargetServiceType(st string) v1.NotificationTargetServiceType {
	switch st {
	case "simple_notification":
		return v1.NotificationTargetServiceTypeSAKURASIMPLENOTICE
	case "eventbus":
		return v1.NotificationTargetServiceTypeSAKURAEVENTBUS
	default:
		return ""
	}
}

func expandNotificationTargetURL(u string) *url.URL {
	if u == "" {
		return nil
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return nil
	}
	return parsed
}
