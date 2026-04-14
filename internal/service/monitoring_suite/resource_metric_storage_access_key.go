// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	monitoringsuiteapi "github.com/sacloud/monitoring-suite-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type metricStorageAccessKeyResource struct {
	client *monitoringsuiteapi.Client
}

var (
	_ resource.Resource                = &metricStorageAccessKeyResource{}
	_ resource.ResourceWithConfigure   = &metricStorageAccessKeyResource{}
	_ resource.ResourceWithImportState = &metricStorageAccessKeyResource{}
)

func NewMetricStorageAccessKeyResource() resource.Resource {
	return &metricStorageAccessKeyResource{}
}

func (r *metricStorageAccessKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_suite_metric_storage_access_key"
}

func (r *metricStorageAccessKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.MonitoringSuiteClient
}

type metricStorageAccessKeyResourceModel struct {
	accessKeyBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *metricStorageAccessKeyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Monitoring Suite Metric Storage Access Key"),
			"description": common.SchemaResourceDescription("Monitoring Suite Metric Storage Access Key"),
			"storage_id": schema.StringAttribute{
				Required:    true,
				Description: "The Metric Storage ID for the Access Key.",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The token of the Access Key.",
			},
			"secret": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The secret of the Access Key.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
		MarkdownDescription: "Manages a Monitoring Suite Metric Storage Access Key.",
	}
}

func (r *metricStorageAccessKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "_", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Import: ID Format Error", "expected import ID format: <storage_id>_<uid>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("storage_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *metricStorageAccessKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan metricStorageAccessKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewMetricsStorageOp(r.client)
	key, err := op.CreateKey(ctx, plan.StorageID.ValueString(), expandOptionalString(plan.Description))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create Metric Storage Access Key: %s", err))
		return
	}

	plan.updateState(plan.StorageID.ValueString(), key.GetUID().String(), key.GetDescription().Value, key.GetToken(), key.GetSecret().String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricStorageAccessKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state metricStorageAccessKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := getMetricsStorageAccessKey(ctx, r.client, state.StorageID.ValueString(), state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	state.updateState(state.StorageID.ValueString(), key.GetUID().String(), key.GetDescription().Value, key.GetToken(), state.Secret.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *metricStorageAccessKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state metricStorageAccessKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewMetricsStorageOp(r.client)
	uid, err := parseUUID(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update: ID Error", fmt.Sprintf("invalid Access Key ID: %s", err))
		return
	}

	key, err := op.UpdateKey(ctx, plan.StorageID.ValueString(), uid, expandOptionalString(plan.Description))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update Metric Storage Access Key[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(plan.StorageID.ValueString(), key.GetUID().String(), key.GetDescription().Value, key.GetToken(), state.Secret.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricStorageAccessKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state metricStorageAccessKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	op := monitoringsuite.NewMetricsStorageOp(r.client)
	uid, err := parseUUID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete: ID Error", fmt.Sprintf("invalid Access Key ID: %s", err))
		return
	}

	if err := op.DeleteKey(ctx, state.StorageID.ValueString(), uid); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete Metric Storage Access Key[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getMetricsStorageAccessKey(ctx context.Context, client *monitoringsuiteapi.Client, storageID string, uid string, state *tfsdk.State, diags *diag.Diagnostics) *monitoringsuiteapi.MetricsStorageAccessKey {
	op := monitoringsuite.NewMetricsStorageOp(client)
	parsedUID, err := parseUUID(uid)
	if err != nil {
		diags.AddError("Read: ID Error", fmt.Sprintf("invalid Access Key ID: %s", err))
		return nil
	}
	key, err := op.ReadKey(ctx, storageID, parsedUID)
	if err != nil {
		if saclient.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Metric Storage Access Key[%s]: %s", uid, err))
		return nil
	}
	return key
}
