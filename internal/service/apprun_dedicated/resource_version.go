// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	ver "github.com/sacloud/apprun-dedicated-api-go/apis/version"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type verResource struct{ resourceClient }

type verResourceModel struct {
	verModel
	RegistryPassword       types.String   `tfsdk:"registry_password"`
	RegistryPasswordAction types.String   `tfsdk:"registry_password_action"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
}

var (
	_ resource.Resource                = &verResource{}
	_ resource.ResourceWithConfigure   = &verResource{}
	_ resource.ResourceWithImportState = &verResource{}
)

func NewVersionResource() resource.Resource { return &verResource{resourceNamed("version")} }

func (r *verResource) Schema(ctx context.Context, _ resource.SchemaRequest, res *resource.SchemaResponse) {
	var m v1.ScalingMode
	modes := common.MapTo(m.AllValues(), common.ToString)
	var rpa v1.RegistryPasswordAction
	actions := common.MapTo(rpa.AllValues(), common.ToString)

	res.Schema = schema.Schema{
		Description: "Manages an AppRun dedicated version",
		Attributes: map[string]schema.Attribute{
			"application_id": func() (attr schema.StringAttribute) {
				attr = common.SchemaResourceId("application").(schema.StringAttribute)
				attr.Required = true
				attr.Computed = false
				attr.Optional = false
				attr.Validators = []validator.String{sacloudvalidator.UUIDValidator}
				attr.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}
				return
			}(),
			"version": schema.Int32Attribute{
				Computed:      true,
				Description:   "The version number",
				PlanModifiers: []planmodifier.Int32{int32planmodifier.RequiresReplace()},
			},
			"cpu": schema.Int64Attribute{
				Required:      true,
				Description:   "The CPU limit in millicores (e.g., 1000 = 1 CPU)",
				Validators:    []validator.Int64{int64validator.Between(100, 64000)},
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"memory": schema.Int64Attribute{
				Required:      true,
				Description:   "The memory limit in megabytes",
				Validators:    []validator.Int64{int64validator.Between(128, 131072)},
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"scaling_mode": schema.StringAttribute{
				Required:      true,
				Description:   "The scaling mode (manual, autoscale)",
				Validators:    []validator.String{stringvalidator.OneOf(modes...)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"fixed_scale": schema.Int32Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: "Number of nodes when scaling mode is `manual`",
				Validators:          []validator.Int32{int32validator.Between(1, 50)},
				PlanModifiers:       []planmodifier.Int32{int32planmodifier.RequiresReplace()},
			},
			"min_scale": schema.Int32Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: "Minimum number of nodes when scaling mode is `autoscale`",
				Validators:          []validator.Int32{int32validator.Between(1, 50)},
				PlanModifiers:       []planmodifier.Int32{int32planmodifier.RequiresReplace()},
			},
			"max_scale": schema.Int32Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: "Maximum number of nodes when scaling mode is `autoscale`",
				Validators:          []validator.Int32{int32validator.Between(1, 50)},
				PlanModifiers:       []planmodifier.Int32{int32planmodifier.RequiresReplace()},
			},
			"scale_in_threshold": schema.Int32Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: "When to scale in when scaling mode is `autoscale`",
				Validators:          []validator.Int32{int32validator.Between(30, 70)},
				PlanModifiers:       []planmodifier.Int32{int32planmodifier.RequiresReplace()},
			},
			"scale_out_threshold": schema.Int32Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: "When to scale out when scaling mode is `autoscale`",
				Validators:          []validator.Int32{int32validator.Between(50, 99)},
				PlanModifiers:       []planmodifier.Int32{int32planmodifier.RequiresReplace()},
			},
			"image": schema.StringAttribute{
				Required:      true,
				Description:   "The container image",
				Validators:    []validator.String{stringvalidator.LengthAtMost(512)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"registry_username": schema.StringAttribute{
				Optional:      true,
				Description:   "Login user name for the container registry",
				Validators:    []validator.String{stringvalidator.LengthAtMost(255)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"registry_password": schema.StringAttribute{
				WriteOnly:     true,
				Optional:      true,
				Description:   "Login password for the container registry",
				Validators:    []validator.String{stringvalidator.LengthAtMost(255)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"registry_password_action": schema.StringAttribute{
				Optional:      true,
				Description:   "Password configuration method",
				Validators:    []validator.String{stringvalidator.OneOf(actions...)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"cmd": schema.ListAttribute{
				Optional:      true,
				ElementType:   types.StringType,
				Description:   "application command line i.e. the command and arguments",
				Validators:    []validator.List{listvalidator.SizeAtMost(20)},
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"created_at": r.schemaCreatedAt(),
			"active_node_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of active nodes.  You might want to ignore_changes this field because it changes from time to time",
			},
			"exposed_ports": schema.SetNestedAttribute{
				Optional:      true,
				Description:   "Ports that the application exposes",
				Validators:    []validator.Set{setvalidator.SizeAtMost(5)},
				PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"target_port": schema.Int32Attribute{
							Required:      true,
							Description:   "The port that the application listens to",
							Validators:    []validator.Int32{int32validator.Between(1, 65535)},
							PlanModifiers: []planmodifier.Int32{int32planmodifier.RequiresReplace()},
						},
						"load_balancer_port": schema.Int32Attribute{
							Optional:      true,
							Description:   "The port that the load balancer listens to, or if when this port is internal",
							Validators:    []validator.Int32{int32validator.Between(1, 65535)},
							PlanModifiers: []planmodifier.Int32{int32planmodifier.RequiresReplace()},
						},
						"use_lets_encrypt": schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Whether the load balancer uses Let's Encrypt (applicable only when `https`)",
							PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
						},
						"host": schema.SetAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Target `Host:` header value (only applicable when `http` or `https`)",
							Validators:          []validator.Set{setvalidator.SizeAtMost(5)},
							PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
						},
						"health_check": schema.SingleNestedAttribute{
							Required:    true,
							Description: "Health check configuration",
							Attributes: map[string]schema.Attribute{
								"path": schema.StringAttribute{
									Required:      true,
									Description:   "Health check endpoint",
									Validators:    []validator.String{stringvalidator.LengthAtMost(200)},
									PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
								},
								"interval": schema.Int32Attribute{
									Required:      true,
									Description:   "Health check intervals in seconds",
									Validators:    []validator.Int32{int32validator.Between(3, 60)},
									PlanModifiers: []planmodifier.Int32{int32planmodifier.RequiresReplace()},
								},
								"timeout": schema.Int32Attribute{
									Required:      true,
									Description:   "Time out in seconds until the health check fails",
									Validators:    []validator.Int32{int32validator.Between(1, 60)},
									PlanModifiers: []planmodifier.Int32{int32planmodifier.RequiresReplace()},
								},
							},
							PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
						},
					},
				},
			},
			"env_vars": schema.SetNestedAttribute{
				Optional:      true,
				Description:   "Environment variables",
				Validators:    []validator.Set{setvalidator.SizeAtMost(50)},
				PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:      true,
							Description:   "Environment variable name",
							Validators:    []validator.String{stringvalidator.LengthBetween(1, 255)},
							PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"value": schema.StringAttribute{
							Optional:      true,
							Description:   "The value.  Omitting this field and set `secret` to true retains old secret value",
							Validators:    []validator.String{stringvalidator.LengthAtMost(4096)},
							PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"secret": schema.BoolAttribute{
							Required:      true,
							Description:   "Whether the value is sensitive",
							PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Delete: true}),
		},
	}
}

func (r *verResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var plan, state verResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	aid, err := plan.appId()

	if err != nil {
		res.Diagnostics.AddError("Create: Invalid Application ID", fmt.Sprintf("failed to parse application ID: %s", err))
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	api := r.api(aid)
	created, err := api.Create(ctx, plan.intoCreate(&state))

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create AppRun Dedicated version: %s", err))
		return
	}

	detail, err := api.Read(ctx, created.Version)

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read created AppRun Dedicated version: %s", err))
		return
	}

	res.Diagnostics.Append(plan.updateState(ctx, detail, aid)...)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *verResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var state verResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	aid, err := state.appId()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Application ID", fmt.Sprintf("failed to parse application ID: %s", err))
		return
	}

	detail, err := r.api(aid).Read(ctx, state.versionNumber())

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated version: %s", err))
		return
	}

	res.Diagnostics.Append(state.updateState(ctx, detail, aid)...)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (r *verResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	// Versions are immutable and cannot be updated
	res.Diagnostics.AddError(
		"Update: Not Supported",
		"AppRun Dedicated versions are immutable. Create a new version instead of updating an existing one.",
	)
}

func (r *verResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	var state verResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	aid, err := state.appId()

	if err != nil {
		res.Diagnostics.AddError("Delete: Invalid Application ID", fmt.Sprintf("failed to parse application ID: %s", err))
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	err = r.api(aid).Delete(ctx, state.versionNumber())

	if err != nil {
		if saclient.IsNotFoundError(err) {
			res.State.RemoveResource(ctx)
			return
		}

		res.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete AppRun Dedicated version: %s", err))
		return
	}
}

func (r *verResource) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	// Import format: application_id/version_number
	parts := strings.Split(req.ID, "/")

	if len(parts) != 2 {
		res.Diagnostics.AddError(
			"Import: Invalid ID",
			fmt.Sprintf("Expected format: application_id/version_number, got: %s", req.ID),
		)
		return
	}

	res.Diagnostics.Append(res.State.SetAttribute(ctx, path.Root("application_id"), parts[0])...)
	res.Diagnostics.Append(res.State.SetAttribute(ctx, path.Root("version"), parts[1])...)
}

func (r *verResource) api(a appID) ver.VersionAPI { return ver.NewVersionOp(r.client, a) }

func (v *verResourceModel) intoCreate(transitional *verResourceModel) (ret ver.CreateParams) {
	ret.CPU = v.CPU.ValueInt64()
	ret.Memory = v.Memory.ValueInt64()
	ret.ScalingMode = v1.ScalingMode(v.ScalingMode.ValueString())
	ret.FixedScale = v.FixedScale.ValueInt32Pointer()
	ret.MinScale = v.MinScale.ValueInt32Pointer()
	ret.MaxScale = v.MinScale.ValueInt32Pointer()
	ret.ScaleInThreshold = v.ScaleInThreshold.ValueInt32Pointer()
	ret.ScaleOutThreshold = v.ScaleOutThreshold.ValueInt32Pointer()
	ret.Image = v.Image.ValueString()
	ret.Cmd = common.TlistToStrings(v.Cmd)
	ret.RegistryUsername = v.RegistryUsername.ValueStringPointer()
	ret.RegistryPassword = transitional.RegistryPassword.ValueStringPointer()
	ret.ExposedPorts = common.MapTo(v.ExposedPorts, exposedPortModel.intoCreate)
	ret.EnvVars = common.MapTo(v.EnvVars, envVarModel.intoCreate)

	// API cannot omit RegistryPasswordAction, but can omit username and password...
	ret.RegistryPasswordAction = v1.RegistryPasswordActionRemove

	if !v.RegistryPasswordAction.IsNull() && !v.RegistryPasswordAction.IsUnknown() {
		ret.RegistryPasswordAction = v1.RegistryPasswordAction(v.RegistryPasswordAction.ValueString())
	}

	return
}
