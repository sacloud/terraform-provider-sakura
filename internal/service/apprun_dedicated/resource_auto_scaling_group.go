// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	asg "github.com/sacloud/apprun-dedicated-api-go/apis/autoscalinggroup"
	lb "github.com/sacloud/apprun-dedicated-api-go/apis/loadbalancer"
	wn "github.com/sacloud/apprun-dedicated-api-go/apis/workernode"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type asgResource struct{ resourceClient }

type asgResourceModel struct {
	asgModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

var (
	_ resource.Resource                = &asgResource{}
	_ resource.ResourceWithConfigure   = &asgResource{}
	_ resource.ResourceWithImportState = &asgResource{}
)

func NewAutoScalingGroupResource() resource.Resource {
	return &asgResource{resourceNamed("auto_scaling_group")}
}

func (r *asgResource) Schema(ctx context.Context, _ resource.SchemaRequest, res *resource.SchemaResponse) {
	nameAttr := r.schemaName(stringvalidator.RegexMatches(
		regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
		"no special characters allowed; alphanumeric and/or hyphens, and underscores",
	))
	nameAttr.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}

	res.Schema = schema.Schema{
		Description: "Manages an AppRun dedicated auto scaling group",
		Attributes: map[string]schema.Attribute{
			"id":         r.schemaID(),
			"cluster_id": r.schemaClusterID(),
			"name":       nameAttr,
			"zone": schema.StringAttribute{
				Required:      true,
				Description:   "The zone name where the auto scaling group will be created",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name_servers": schema.ListAttribute{
				Optional:      true,
				ElementType:   types.StringType,
				Description:   "The name servers for the auto scaling group (ORDER MATTERS)",
				Validators:    []validator.List{listvalidator.SizeAtMost(3)},
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"worker_service_class_path": schema.StringAttribute{
				Required:      true,
				Description:   "The worker service class path",
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 255)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"min_nodes": schema.Int32Attribute{
				Required:      true,
				Description:   "Minimum number of nodes",
				Validators:    []validator.Int32{int32validator.Between(1, 10)},
				PlanModifiers: []planmodifier.Int32{int32planmodifier.RequiresReplace()},
			},
			"max_nodes": schema.Int32Attribute{
				Required:      true,
				Description:   "Maximum number of nodes",
				Validators:    []validator.Int32{int32validator.Between(1, 10)},
				PlanModifiers: []planmodifier.Int32{int32planmodifier.RequiresReplace()},
			},
			"current_nodes": schema.Int32Attribute{
				Computed:    true,
				Description: "The current number of nodes. You might want to ignore_changes this field because it changes from time to time",
			},
			"deleting": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the auto scaling group is being deleted",
			},
			"interfaces": schema.SetNestedAttribute{
				Required:      true,
				Description:   "The network interfaces for the nodes",
				Validators:    []validator.Set{setvalidator.SizeBetween(1, 5)},
				PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"interface_index": schema.Int32Attribute{
							Required:      true,
							Description:   "The interface index",
							Validators:    []validator.Int32{int32validator.AtLeast(0)},
							PlanModifiers: []planmodifier.Int32{int32planmodifier.RequiresReplace()},
						},
						"upstream": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The upstream switch id, or `shared` to use shared segment",
							PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"ip_pool": schema.SetNestedAttribute{
							Optional:      true,
							Description:   "The IP pool for the interface.  Must omit when upstream is `shared`.  Mandatory otherwise.",
							Validators:    []validator.Set{setvalidator.SizeAtMost(20)},
							PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()},
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"start": schema.StringAttribute{
										Required:    true,
										Description: "The start IP address of the range",
										Validators: []validator.String{stringvalidator.RegexMatches(
											regexp.MustCompile(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`),
											"must be an IPv4 address",
										)},
										PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
									},
									"end": schema.StringAttribute{
										Required:    true,
										Description: "The end IP address of the range",
										Validators: []validator.String{stringvalidator.RegexMatches(
											regexp.MustCompile(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`),
											"must be an IPv4 address",
										)},
										PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
									},
								},
							},
						},
						"netmask_len": schema.Int32Attribute{
							Optional:            true,
							MarkdownDescription: "The netmask length.  Must omit when upstream is `shared`.  Mandatory otherwise.",
							Validators:          []validator.Int32{int32validator.Between(8, 29)},
							PlanModifiers:       []planmodifier.Int32{int32planmodifier.RequiresReplace()},
						},
						"default_gateway": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The default gateway.  Makes sense only when upstream is not `shared`",
							PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"packet_filter_id": schema.StringAttribute{
							Optional:      true,
							Description:   "The packet filter ID",
							PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"connects_to_lb": schema.BoolAttribute{
							Required:      true,
							Description:   "Whether the interface connects to the load balancer",
							PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Delete: true}),
		},
	}
}

func (r *asgResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var plan asgResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if res.Diagnostics.HasError() {
		return
	}

	cid, err := plan.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Create: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	params, diag := plan.intoCreate()
	res.Diagnostics.Append(diag...)

	if res.Diagnostics.HasError() {
		return
	}

	api := r.api(cid)
	created, err := api.Create(ctx, params)

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create AppRun Dedicated auto scaling group: %s", err))
		return
	}

	detail, err := api.Read(ctx, created.AutoScalingGroupID)

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read created AppRun Dedicated auto scaling group: %s", err))
		return
	}

	res.Diagnostics.Append(plan.updateState(ctx, detail, cid)...)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *asgResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var state asgResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	cid, id, err := state.ids()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid IDs", fmt.Sprintf("failed to parse IDs: %s", err))
		return
	}

	detail, err := r.api(cid).Read(ctx, id)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated auto scaling group: %s", err))
		return
	}

	res.Diagnostics.Append(state.updateState(ctx, detail, cid)...)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (r *asgResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	// Auto scaling groups are immutable and cannot be updated
	res.Diagnostics.AddError(
		"Update: Not Supported",
		"AppRun Dedicated auto scaling groups are immutable. Create a new auto scaling group instead of updating an existing one.",
	)
}

func (r *asgResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	var state asgResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	// THIS OPERATION TAKES LOOONG TIME
	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	err := state.waitDeleted(ctx, r)

	if err != nil {
		if saclient.IsNotFoundError(err) {
			res.State.RemoveResource(ctx)
			return
		}

		res.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete AppRun Dedicated auto scaling group: %s", err))
		return
	}
}

func (r *asgResource) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	// Import format: cluster_id/auto_scaling_group_id
	parts := strings.Split(req.ID, "/")

	if len(parts) != 2 {
		res.Diagnostics.AddError(
			"Import: Invalid ID",
			fmt.Sprintf("Expected format: cluster_id/auto_scaling_group_id, got: %s", req.ID),
		)
		return
	}

	res.Diagnostics.Append(res.State.SetAttribute(ctx, path.Root("cluster_id"), parts[0])...)
	res.Diagnostics.Append(res.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *asgResource) api(c clusterID) asg.AutoScalingGroupAPI {
	return asg.NewAutoScalingGroupOp(r.client, c)
}

func (r *asgResourceModel) ids() (c clusterID, a asgID, e error) {
	c, e = intoUUID[clusterID](r.ClusterID)

	if e != nil {
		return
	}

	a, e = intoUUID[asgID](r.ID)

	return
}

func (r *asgResourceModel) waitDeleted(ctx context.Context, client *asgResource) (err error) {
	// An auto scaling group can have load balancer nodes and worker nodes
	// They have to be provisioned before deleted
	wg := sync.WaitGroup{}
	ch := make(chan error, 2)
	c, a, err := r.ids()

	if err != nil {
		return err
	}

	wnAPI := wn.NewWorkerNodeOp(client.client, c, a)
	lbAPI := lb.NewLoadBalancerOp(client.client, c, a)

	tflog.Info(ctx, "waiting for ASG provisioning", map[string]any{"id": uuid.UUID(a).String()})
	wg.Go(func() { ch <- deleteLBs(ctx, lbAPI) })
	wg.Go(func() { ch <- deleteWNs(ctx, wnAPI) })
	go func() {
		defer close(ch)
		wg.Wait()
	}()
	for e := range ch {
		err = errors.Join(err, e)
	}

	if err != nil {
		return err
	}

	err = client.api(c).Delete(ctx, a)

	if saclient.IsNotFoundError(err) {
		return nil // no asg no error
	}

	if err != nil {
		return err
	}

	return drainASG(ctx, client.api(c), a)
}

func drainASG(ctx context.Context, api asg.AutoScalingGroupAPI, id asgID) error {
	t := time.NewTicker(13 * time.Second)
	defer t.Stop()

	for {
		_, err := api.Read(ctx, id)

		if saclient.IsNotFoundError(err) {
			tflog.Debug(ctx, "ASG deleted", map[string]any{"id": uuid.UUID(id).String()})
			return nil // no asg no problem
		}

		if err != nil {
			return err
		}

		tflog.Debug(ctx, "ASG deleting", map[string]any{"id": uuid.UUID(id).String()})

		select {
		case <-t.C:
			continue

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
