// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	app "github.com/sacloud/apprun-dedicated-api-go/apis/application"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type appResource struct{ resourceClient }

type appResourceModel struct {
	appModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

var (
	_ resource.Resource                = &appResource{}
	_ resource.ResourceWithConfigure   = &appResource{}
	_ resource.ResourceWithImportState = &appResource{}
)

func NewAppResource() resource.Resource { return &appResource{resourceNamed("application")} }

func (r *appResource) Schema(ctx context.Context, _ resource.SchemaRequest, res *resource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "Manages an AppRun dedicated application",
		Attributes: map[string]schema.Attribute{
			"id":         r.schemaID(),
			"cluster_id": r.schemaClusterID(),
			"name": r.schemaName(stringvalidator.RegexMatches(
				regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
				"no special characters allowed; alphanumeric and/or hyphens and underscores",
			)),
			"cluster_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the cluster",
			},
			"active_version": schema.Int32Attribute{
				Optional:    true,
				Description: "The active version of the application",
			},
			"desired_count": schema.Int32Attribute{
				Computed:    true,
				Description: "The desired count of the application",
			},
			"scaling_cooldown_seconds": schema.Int32Attribute{
				Computed:    true,
				Description: "The scaling cooldown seconds of the application",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var plan appResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if res.Diagnostics.HasError() {
		return
	}

	clusterID, err := plan.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Create: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	api := r.api()
	created, err := api.Create(ctx, plan.intoCreate(), clusterID)

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create AppRun Dedicated application: %s", err))
		return
	}

	detail, err := api.Read(ctx, created.ApplicationID)

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read created AppRun Dedicated application: %s", err))
		return
	}

	plan.updateState(detail)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var state appResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	appID, err := state.appId()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Application ID", fmt.Sprintf("failed to parse application ID: %s", err))
		return
	}

	api := r.api()

	detail, err := api.Read(ctx, appID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated application: %s", err))
		return
	}

	state.updateState(detail)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (r *appResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	var plan appResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if res.Diagnostics.HasError() {
		return
	}

	appID, err := plan.appId()

	if err != nil {
		res.Diagnostics.AddError("Update: Invalid Application ID", fmt.Sprintf("failed to parse application ID: %s", err))
		return
	}

	api := r.api()

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	err = api.Update(ctx, appID, plan.intoUpdate())

	if err != nil {
		res.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update AppRun Dedicated application: %s", err))
		return
	}

	detail, err := api.Read(ctx, appID)

	if err != nil {
		res.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read AppRun Dedicated application: %s", err))
		return
	}

	plan.updateState(detail)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *appResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	var state appResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	appID, err := state.appId()

	if err != nil {
		res.Diagnostics.AddError("Delete: Invalid Application ID", fmt.Sprintf("failed to parse application ID: %s", err))
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	err = r.api().Delete(ctx, appID)

	if err != nil {
		if saclient.IsNotFoundError(err) {
			res.State.RemoveResource(ctx)
			return
		}
		res.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete AppRun Dedicated application: %s", err))
		return
	}
}

func (r *appResource) api() *app.ApplicationOp { return app.NewApplicationOp(r.client) }
func (c *appResourceModel) intoCreate() string { return c.Name.ValueString() }

func (c *appResourceModel) intoUpdate() *int32 {
	switch {
	case c.ActiveVersion.IsUnknown():
		return (*int32)(nil)
	case c.ActiveVersion.IsNull():
		return (*int32)(nil)
	default:
		return c.ActiveVersion.ValueInt32Pointer()
	}
}
