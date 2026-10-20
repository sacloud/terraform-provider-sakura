// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	app "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/application"
	v1 "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/v1"
	ver "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/version"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type activeVerResource struct{ resourceClient }

var (
	_ resource.Resource                   = &activeVerResource{}
	_ resource.ResourceWithConfigure      = &activeVerResource{}
	_ resource.ResourceWithImportState    = &activeVerResource{}
	_ resource.ResourceWithValidateConfig = &activeVerResource{}
)

func NewActiveVersionResource() resource.Resource {
	return &activeVerResource{resourceNamed("active_version")}
}

func (r *activeVerResource) Schema(ctx context.Context, _ resource.SchemaRequest, res *resource.SchemaResponse) {
	var m v1.ScalingMode
	modes := common.MapTo(m.AllValues(), common.ToString)
	var rpa v1.RegistryPasswordAction
	actions := common.MapTo(rpa.AllValues(), common.ToString)

	res.Schema = schema.Schema{
		MarkdownDescription: "Manages the active version of an AppRun dedicated application.  " +
			"Every change to the version configuration creates a new version and activates it; previous versions are kept.  The number of the active version is `version`.  " +
			"Create and update wait until the previous version has left every node, not until the new one is running; raise `timeouts.update` for a large `fixed_scale`.\n\n" +
			"Unlike `sakura_apprun_dedicated_version`, this resource represents whichever version is active, so a version activated outside of Terraform is followed on refresh.  " +
			"An apply whose plan predates such an activation is refused; the check is best effort and does not cover an activation during the apply itself.  " +
			"Do not combine it with `sakura_apprun_dedicated_application.active_version` (both would activate versions) " +
			"or `sakura_apprun_dedicated_version` (its version cannot be deleted while this resource keeps it active) on the same application.\n\n" +
			"Import with the application ID; the application must have an active version.  The first plan after an import may order `env_vars` and `secret_vars` differently; list them in the order of the state.  " +
			"Destroying the resource, or removing it from the configuration, deactivates the application and waits until its containers have stopped; use `terraform state rm` to stop managing the active version without deactivating it.  " +
			"After a deactivation outside of Terraform, the next apply creates a new version from the configuration, `image` included.\n\n" +
			"To deploy the image outside of Terraform (e.g. from a CI/CD pipeline), add `image` to `ignore_changes`; " +
			"it must still be set for the first apply, and later versions are created with the image currently deployed:\n\n" +
			"```terraform\n" +
			"resource \"sakura_apprun_dedicated_active_version\" \"main\" {\n" +
			"  # ...\n" +
			"  lifecycle {\n" +
			"    ignore_changes = [image]\n" +
			"  }\n" +
			"}\n" +
			"```",
		Attributes: map[string]schema.Attribute{
			"id": func() (attr schema.StringAttribute) {
				attr = r.schemaID()
				attr.MarkdownDescription = "The ID of the application, the same as `application_id`"
				return
			}(),
			"application_id": r.schemaAppID(),
			"version": schema.Int32Attribute{
				Computed:    true,
				Description: "The number of the active version.  Increases whenever a new version is created",
			},
			"cpu": schema.Int64Attribute{
				Required:    true,
				Description: desc.Sprintf("The CPU limit in millicores (e.g., 1000 = 1 CPU). %s", desc.Range(100, 64000)),
				Validators:  []validator.Int64{int64validator.Between(100, 64000)},
			},
			"memory": schema.Int64Attribute{
				Required:    true,
				Description: desc.Sprintf("The memory limit in megabytes. %s", desc.Range(128, 131072)),
				Validators:  []validator.Int64{int64validator.Between(128, 131072)},
			},
			"scaling_mode": schema.StringAttribute{
				Required:    true,
				Description: "The scaling mode (manual, autoscale)",
				Validators:  []validator.String{stringvalidator.OneOf(modes...)},
			},
			"fixed_scale": schema.Int32Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: desc.Sprintf("Number of nodes when scaling mode is `manual`. %s", desc.Range(1, 50)),
				Validators:          []validator.Int32{int32validator.Between(1, 50)},
			},
			"min_scale": schema.Int32Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: desc.Sprintf("Minimum number of nodes when scaling mode is `autoscale`. %s", desc.Range(1, 50)),
				Validators:          []validator.Int32{int32validator.Between(1, 50)},
			},
			"max_scale": schema.Int32Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: desc.Sprintf("Maximum number of nodes when scaling mode is `autoscale`. %s", desc.Range(1, 50)),
				Validators:          []validator.Int32{int32validator.Between(1, 50)},
			},
			"scale_in_threshold": schema.Int32Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: desc.Sprintf("When to scale in when scaling mode is `autoscale`. %s", desc.Range(30, 70)),
				Validators:          []validator.Int32{int32validator.Between(30, 70)},
			},
			"scale_out_threshold": schema.Int32Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: desc.Sprintf("When to scale out when scaling mode is `autoscale`. %s", desc.Range(50, 99)),
				Validators:          []validator.Int32{int32validator.Between(50, 99)},
			},
			"image": schema.StringAttribute{
				Required:    true,
				Description: "The container image, e.g. `nginx:latest`",
				Validators:  []validator.String{stringvalidator.LengthAtMost(512)},
			},
			"registry_username": schema.StringAttribute{
				Optional:    true,
				Description: "Login user name for the container registry",
				Validators:  []validator.String{stringvalidator.LengthAtMost(255)},
			},
			"registry_password": schema.StringAttribute{
				WriteOnly:           true,
				Optional:            true,
				MarkdownDescription: "Login password for the container registry.  A change of this value alone makes no plan; change it together with another attribute, e.g. `registry_password_action`",
				Validators:          []validator.String{stringvalidator.LengthAtMost(255)},
			},
			"registry_password_action": schema.StringAttribute{
				Optional:    true,
				Description: desc.Sprintf("Password configuration method. This must be one of [%s].  Note that `terraform import` cannot restore this field", actions),
				Validators:  []validator.String{stringvalidator.OneOf(actions...)},
			},
			"cmd": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "application command line i.e. the command and arguments",
				Validators:  []validator.List{listvalidator.SizeAtMost(20)},
			},
			// no UseStateForUnknown: every update creates a new version
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation timestamp of the active version",
			},
			"active_node_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of nodes running the active version",
			},
			"exposed_ports": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Ports that the application exposes",
				Validators:  []validator.List{listvalidator.SizeAtMost(5)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"target_port": schema.Int32Attribute{
							Required:    true,
							Description: desc.Sprintf("The port that the application listens to. %s", desc.Range(1, 65535)),
							Validators:  []validator.Int32{int32validator.Between(1, 65535)},
						},
						"lb_port": schema.Int32Attribute{
							Optional:            true,
							MarkdownDescription: desc.Sprintf("The port that the load balancer listens to.  Explicitly set it to `null` when you want to disconnect from the load balancer. %s", desc.Range(1, 65535)),
							Validators:          []validator.Int32{int32validator.Between(1, 65535)},
						},
						"use_lets_encrypt": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Whether the load balancer uses Let's Encrypt (applicable only when `https`)",
						},
						"host": schema.SetAttribute{
							Optional:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Target `Host:` header value (only applicable when `http` or `https`)",
							Validators:          []validator.Set{setvalidator.SizeAtMost(5)},
						},
						"health_check": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "Health check configuration",
							Attributes: map[string]schema.Attribute{
								"path": schema.StringAttribute{
									Required:    true,
									Description: "Health check endpoint",
									Validators:  []validator.String{stringvalidator.LengthAtMost(200)},
								},
								"interval_seconds": schema.Int32Attribute{
									Required:    true,
									Description: desc.Sprintf("Health check intervals in seconds. %s", desc.Range(3, 60)),
									Validators:  []validator.Int32{int32validator.Between(3, 60)},
								},
								"timeout_seconds": schema.Int32Attribute{
									Required:    true,
									Description: desc.Sprintf("Time out in seconds until the health check fails. %s", desc.Range(1, 60)),
									Validators:  []validator.Int32{int32validator.Between(1, 60)},
								},
							},
						},
					},
				},
			},
			"env_vars": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Environment variables.  Use `secret_vars` for secrets",
				Validators:  []validator.List{listvalidator.SizeAtMost(50)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "Environment variable name",
							Validators:  []validator.String{stringvalidator.LengthBetween(1, 255)},
						},
						"value": schema.StringAttribute{
							Optional:    true,
							Description: "The value.  Use `secret_vars` for secrets.  Omitting this field with the deprecated `secret = true` retains the previous version's secret value",
							Validators:  []validator.String{stringvalidator.LengthAtMost(4096)},
						},
						"secret": schema.BoolAttribute{
							Optional:           true,
							Computed:           true,
							Default:            booldefault.StaticBool(false),
							Description:        "Whether the value is sensitive",
							DeprecationMessage: "The \"secret\" attribute is deprecated and will be removed in a future version. Omit it (it defaults to false) and put secrets in secret_vars instead.",
						},
					},
				},
			},
			"secret_vars": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Secret environment variables.  Unlike `env_vars`, values are write-only and never land in the state.  `terraform import` puts every secret here; declare each as `{ key = \"...\" }` alone to adopt the imported version as is, and set `value_wo` only when a new version with a new value is wanted",
				Validators:  []validator.List{listvalidator.SizeAtMost(50)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:    true,
							Description: "Environment variable name",
							Validators:  []validator.String{stringvalidator.LengthBetween(1, 255)},
						},
						"value_wo": schema.StringAttribute{
							WriteOnly:   true,
							Optional:    true,
							Description: "The value, write-only.  Omitting this field retains the previous version's secret value.  Must be set together with `value_wo_version`",
							Validators: []validator.String{
								stringvalidator.LengthAtMost(4096),
								stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("value_wo_version")),
							},
						},
						"value_wo_version": schema.Int32Attribute{
							Optional:    true,
							Description: "The version of the `value_wo` field.  This value must be greater than 0 when set.  Terraform cannot detect changes of write-only values, so increment this to create a new application version with the new `value_wo`.  Note that `terraform import` cannot restore this field",
							Validators: []validator.Int32{
								int32validator.AtLeast(1),
								int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("value_wo")),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func (r *activeVerResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, res *resource.ValidateConfigResponse) {
	validateVersionConfig(ctx, req, res)
}

func (r *activeVerResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var plan, config activeVerResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	res.Diagnostics.Append(req.Config.Get(ctx, &config)...)

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

	params, diag := plan.intoCreate(&config.verResourceModel)
	res.Diagnostics.Append(diag...)

	if res.Diagnostics.HasError() {
		return
	}

	detail, err := r.activate(ctx, params, aid)

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", err.Error())
		return
	}

	res.Diagnostics.Append(plan.updateState(ctx, detail, aid)...)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *activeVerResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var state activeVerResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	// after import only id is set, hence not state.appId()
	aid, err := intoUUID[appID](state.ID)

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid Application ID", fmt.Sprintf("failed to parse application ID: %s", err))
		return
	}

	detail, err := r.appAPI().Read(ctx, aid)

	if err != nil {
		if saclient.IsNotFoundError(err) {
			res.State.RemoveResource(ctx)
			return
		}

		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated application: %s", err))
		return
	}

	if detail.ActiveVersion == nil {
		res.State.RemoveResource(ctx)
		return
	}

	version, err := r.verAPI(aid).Read(ctx, v1.ApplicationVersionNumber(*detail.ActiveVersion))

	if err != nil {
		if saclient.IsNotFoundError(err) {
			res.State.RemoveResource(ctx)
			return
		}

		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated version: %s", err))
		return
	}

	state.dropMissingVars(version)
	res.Diagnostics.Append(state.updateState(ctx, version, aid)...)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (r *activeVerResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	var plan, state, config activeVerResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)
	res.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if res.Diagnostics.HasError() {
		return
	}

	aid, err := plan.appId()

	if err != nil {
		res.Diagnostics.AddError("Update: Invalid Application ID", fmt.Sprintf("failed to parse application ID: %s", err))
		return
	}

	// a timeouts-only change must not create a version: compare the plan, with timeouts
	// and the computed attributes taken from the state, against the prior state
	same := plan
	same.Timeouts = state.Timeouts
	same.Version = state.Version
	same.CreatedAt = state.CreatedAt
	same.ActiveNodeCount = state.ActiveNodeCount

	samePlan := tfsdk.Plan{Schema: req.Plan.Schema}
	res.Diagnostics.Append(samePlan.Set(ctx, &same)...)

	if res.Diagnostics.HasError() {
		return
	}

	if samePlan.Raw.Equal(req.State.Raw) {
		plan.Version = state.Version
		plan.CreatedAt = state.CreatedAt
		plan.ActiveNodeCount = state.ActiveNodeCount
		res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	// refuse to overwrite a version activated since the plan was made
	detail, err := r.appAPI().Read(ctx, aid)

	if err != nil {
		res.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read AppRun Dedicated application: %s", err))
		return
	}

	if detail.ActiveVersion == nil || *detail.ActiveVersion != state.Version.ValueInt32() {
		res.Diagnostics.AddError(
			"Update: Active Version Changed",
			fmt.Sprintf("the active version of the application changed outside of Terraform (expected %d, got %s). Run terraform plan again to refresh the state before applying.", state.Version.ValueInt32(), types.Int32PointerValue(detail.ActiveVersion)),
		)
		return
	}

	params, diag := plan.intoCreate(&config.verResourceModel)
	res.Diagnostics.Append(diag...)

	if res.Diagnostics.HasError() {
		return
	}

	version, err := r.activate(ctx, params, aid)

	if err != nil {
		res.Diagnostics.AddError("Update: API Error", err.Error())
		return
	}

	res.Diagnostics.Append(plan.updateState(ctx, version, aid)...)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *activeVerResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	var state activeVerResourceModel
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

	// deactivate only; versions are left intact
	api := r.appAPI()
	err = api.Update(ctx, aid, nil)

	if err != nil {
		if saclient.IsNotFoundError(err) {
			res.State.RemoveResource(ctx)
			return
		}

		res.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to deactivate AppRun Dedicated version: %s", err))
		return
	}

	// the API refuses to delete an application whose containers still run
	err = waitForActiveVersion(ctx, api, aid, nil)

	if err != nil {
		if saclient.IsNotFoundError(err) {
			res.State.RemoveResource(ctx)
			return
		}

		res.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to wait for AppRun Dedicated containers to stop: %s", err))
		return
	}
}

func (r *activeVerResource) appAPI() *app.ApplicationOp    { return app.NewApplicationOp(r.client) }
func (r *activeVerResource) verAPI(a appID) ver.VersionAPI { return ver.NewVersionOp(r.client, a) }

// creates a version, activates it and waits until the previous version has left the nodes
func (r *activeVerResource) activate(ctx context.Context, params ver.CreateParams, aid appID) (*ver.VersionDetail, error) {
	verAPI := r.verAPI(aid)
	created, err := verAPI.Create(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("failed to create AppRun Dedicated version: %w", err)
	}

	version := common.ToInt32(created.Version)
	appAPI := r.appAPI()
	err = appAPI.Update(ctx, aid, &version)

	if err != nil {
		return nil, fmt.Errorf("failed to activate AppRun Dedicated version %d: %w", version, err)
	}

	// the version is active from here on: say so in the errors, so that it can be imported
	err = waitForActiveVersion(ctx, appAPI, aid, &version)

	if err != nil {
		return nil, fmt.Errorf("failed to wait for AppRun Dedicated version %d to become active (it is activated; run terraform import with the application ID to adopt it): %w", version, err)
	}

	detail, err := verAPI.Read(ctx, created.Version)

	if err != nil {
		return nil, fmt.Errorf("failed to read activated AppRun Dedicated version %d (it is activated; run terraform import with the application ID to adopt it): %w", version, err)
	}

	return detail, nil
}

func waitForActiveVersion(ctx context.Context, api app.ApplicationAPI, aid appID, version *int32) error {
	timedOut := func() error {
		if version == nil {
			return fmt.Errorf("timed out waiting for the containers to stop: %w", ctx.Err())
		}

		return fmt.Errorf("timed out waiting for version %d to be placed: %w", *version, ctx.Err())
	}

	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	for {
		// a done context would fail the call with the raw context error
		if ctx.Err() != nil {
			return timedOut()
		}

		placements, err := api.Containers(ctx, aid)

		if err != nil {
			return err
		}

		if isPlacementConverged(placements, version) {
			return nil
		}

		select {
		case <-t.C:
			continue

		case <-ctx.Done():
			return timedOut()
		}
	}
}

// no node has a container of another version, placed or desired; nil: no container at all
func isPlacementConverged(placements []app.Placement, version *int32) bool {
	for _, p := range placements {
		for _, c := range p.Desired.Containers {
			if version == nil || c.ApplicationVersion != int64(*version) {
				return false
			}
		}

		for _, c := range p.ContainersStats.Containers {
			if version == nil || c.ApplicationVersion != int64(*version) {
				return false
			}
		}
	}

	return true
}
