// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	lb "github.com/sacloud/apprun-dedicated-api-go/apis/loadbalancer"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type lbResource struct{ resourceClient }

type lbResourceModel struct {
	lbModel
	ClusterID          types.String   `tfsdk:"cluster_id"`
	AutoScalingGroupID types.String   `tfsdk:"auto_scaling_group_id"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

var (
	_ resource.Resource                = &lbResource{}
	_ resource.ResourceWithConfigure   = &lbResource{}
	_ resource.ResourceWithImportState = &lbResource{}
)

func NewLoadBalancerResource() resource.Resource {
	return &lbResource{resourceNamed("lb")}
}

func (r *lbResource) Schema(ctx context.Context, _ resource.SchemaRequest, res *resource.SchemaResponse) {
	nameAttr := r.schemaName(stringvalidator.RegexMatches(
		regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
		"no special characters allowed; alphanumeric and/or hyphens, and underscores",
	))
	nameAttr.PlanModifiers = []planmodifier.String{stringplanmodifier.RequiresReplace()}

	res.Schema = schema.Schema{
		Description: "Manages an AppRun dedicated load balancer",
		Attributes: map[string]schema.Attribute{
			"id":                    r.schemaID(),
			"cluster_id":            r.schemaClusterID(),
			"auto_scaling_group_id": r.schemaASGID(),
			"name":                  nameAttr,
			"service_class_path": schema.StringAttribute{
				Required:      true,
				Description:   "The service class path for the load balancer",
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 255)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name_servers": schema.ListAttribute{
				Optional:      true,
				ElementType:   types.StringType,
				Description:   "The name servers for the load balancer (ORDER MATTERS)",
				Validators:    []validator.List{listvalidator.SizeAtMost(3)},
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"interfaces": schema.SetNestedAttribute{
				Required:      true,
				Description:   "The network interfaces for the load balancer",
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
						"netmask": schema.Int32Attribute{
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
						"vip": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The VIP address. Makes sense only when upstream is not `shared`",
							PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
							Validators: []validator.String{stringvalidator.RegexMatches(
								regexp.MustCompile(`^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$`),
								"must be an IPv4 address",
							)},
						},
						"virtual_router_id": schema.Int32Attribute{
							Optional:            true,
							MarkdownDescription: "The virtual router ID. Makes sense only when upstream is not `shared`",
							Validators:          []validator.Int32{int32validator.Between(1, 255)},
							PlanModifiers:       []planmodifier.Int32{int32planmodifier.RequiresReplace()},
						},
						"packet_filter_id": schema.StringAttribute{
							Optional:      true,
							Description:   "The packet filter ID",
							PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
					},
				},
			},
			"created":  r.schemaCreatedAt(),
			"deleting": schema.BoolAttribute{Computed: true, Description: "Whether the load balancer is being deleted"},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Delete: true}),
		},
	}
}

func (r *lbResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var plan lbResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if res.Diagnostics.HasError() {
		return
	}

	cid, err := plan.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Create: Invalid Cluster ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	asgID, err := plan.asgID()

	if err != nil {
		res.Diagnostics.AddError("Create: Invalid Auto Scaling Group ID", fmt.Sprintf("failed to parse auto scaling group ID: %s", err))
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	api := r.api(cid, asgID)
	params, diag := plan.intoCreate()
	res.Diagnostics.Append(diag...)

	if res.Diagnostics.HasError() {
		return
	}

	created, err := api.Create(ctx, params)

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create AppRun Dedicated load balancer: %s", err))
		return
	}

	detail, err := api.Read(ctx, created.LoadBalancerID)

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read created AppRun Dedicated load balancer: %s", err))
		return
	}

	res.Diagnostics.Append(plan.updateState(ctx, detail)...)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *lbResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var state lbResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	cid, asgID, lbID, err := state.ids()

	if err != nil {
		res.Diagnostics.AddError("Read: Invalid IDs", fmt.Sprintf("failed to parse IDs: %s", err))
		return
	}

	detail, err := r.api(cid, asgID).Read(ctx, lbID)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated load balancer: %s", err))
		return
	}

	res.Diagnostics.Append(state.updateState(ctx, detail)...)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (r *lbResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	// Load balancers are immutable and cannot be updated
	res.Diagnostics.AddError(
		"Update: Not Supported",
		"AppRun Dedicated load balancers are immutable. Create a new load balancer instead of updating an existing one.",
	)
}

func (r *lbResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	var state lbResourceModel
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

		res.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete AppRun Dedicated load balancer: %s", err))
		return
	}
}

func (r *lbResource) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	// Import format: cluster_id/auto_scaling_group_id/lb_id
	parts := strings.Split(req.ID, "/")

	if len(parts) != 3 {
		res.Diagnostics.AddError(
			"Import: Invalid ID",
			fmt.Sprintf("Expected format: cluster_id/auto_scaling_group_id/lb_id, got: %s", req.ID),
		)
		return
	}

	res.Diagnostics.Append(res.State.SetAttribute(ctx, path.Root("cluster_id"), parts[0])...)
	res.Diagnostics.Append(res.State.SetAttribute(ctx, path.Root("auto_scaling_group_id"), parts[1])...)
	res.Diagnostics.Append(res.State.SetAttribute(ctx, path.Root("id"), parts[2])...)
}

func (r *lbResource) api(c clusterID, a asgID) lb.LoadBalancerAPI {
	return lb.NewLoadBalancerOp(r.client, c, a)
}

func (r *lbResourceModel) ids() (c clusterID, a asgID, l lbID, e error) {
	c, e = r.clusterID()

	if e != nil {
		return
	}

	a, e = r.asgID()

	if e != nil {
		return
	}

	l, e = r.lbID()

	return
}

func (r *lbResourceModel) clusterID() (clusterID, error) { return intoUUID[clusterID](r.ClusterID) }
func (r *lbResourceModel) asgID() (asgID, error)         { return intoUUID[asgID](r.AutoScalingGroupID) }

func (r *lbResourceModel) waitDeleted(ctx context.Context, client *lbResource) (err error) {
	c, a, l, err := r.ids()

	if err != nil {
		return err
	}

	api := client.api(c, a)

	// Wait for nodes to be provisioned before deleting
	err = provisionLBNodes(ctx, api, l)

	if err != nil {
		return err
	}

	return drainLB(ctx, api, l)
}

func provisionLBNodes(ctx context.Context, api lb.LoadBalancerAPI, id lbID) error {
	t := time.NewTicker(7 * time.Second)
	defer t.Stop()

	for {
		ok, err := provisionLBsInternalNodes(ctx, api, id)

		if saclient.IsNotFoundError(err) {
			return nil // no lb no problem
		}

		if err != nil {
			return err
		}

		if ok {
			return nil
		}

		select {
		case <-t.C:
			continue

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func drainLB(ctx context.Context, api lb.LoadBalancerAPI, id lbID) error {
	t := time.NewTicker(13 * time.Second)
	defer t.Stop()

	for {
		ok, err := waitLBsInternalLB(ctx, api, id)

		if saclient.IsNotFoundError(err) {
			return nil // no lb no problem
		}

		if err != nil {
			return err
		}

		if ok {
			return nil
		}

		tflog.Debug(ctx, "ASG LB deleting", map[string]any{"id": uuid.UUID(id).String()})

		select {
		case <-t.C:
			continue

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
