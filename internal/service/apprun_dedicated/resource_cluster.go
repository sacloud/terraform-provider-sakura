// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/apprun-dedicated-api-go/apis/cluster"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type clusterResource struct{ resourceClient }

type clusterResourceModel struct {
	clusterModel

	LetsEncryptEmail types.String   `tfsdk:"lets_encrypt_email"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

var (
	_ resource.Resource                = &clusterResource{}
	_ resource.ResourceWithConfigure   = &clusterResource{}
	_ resource.ResourceWithImportState = &clusterResource{}
)

func NewClusterResource() resource.Resource { return &clusterResource{resourceNamed("cluster")} }

var reservedPorts = []int32{
	5950, 5951, 5952, 5953, 5954, 5955, 5956, 5957, 5958, 5959,
}

func (r *clusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, res *resource.SchemaResponse) {
	id := r.schemaID()

	name := r.schemaName(stringvalidator.RegexMatches(
		regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
		"no special characters allowed; alphanumeric and/or hyphens and underscores",
	))

	email := schema.StringAttribute{
		Optional:      true,
		Description:   "Let'sEncrypt registation email address",
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}

	spid := schema.StringAttribute{
		Required:    true,
		Description: "The service principal ID. This is the principal that invokes the application",
		Validators: []validator.String{
			stringvalidator.LengthBetween(12, 12),
			sacloudvalidator.SakuraIDValidator(),
		},
	}

	portno := schema.Int32Attribute{
		Required:    true,
		Description: "The port number where the cluster listens for requests",
		Validators: []validator.Int32{
			int32validator.Between(1, 65535),
			int32validator.NoneOf(reservedPorts...),
		},
	}

	var p v1.CreateLoadBalancerPortProtocol
	protocols := common.MapTo(p.AllValues(), common.ToString)
	protocol := schema.StringAttribute{
		Required:            true,
		MarkdownDescription: "Either `http`, `https`, or `tcp`",
		Validators:          []validator.String{stringvalidator.OneOf(protocols...)},
	}

	nested := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"port":     portno,
			"protocol": protocol,
		},
	}

	ports := schema.SetNestedAttribute{
		Optional:      true,
		Description:   "The list of ports that the cluster listens on (max 5)",
		NestedObject:  nested,
		Validators:    []validator.Set{setvalidator.SizeAtMost(5)},
		PlanModifiers: []planmodifier.Set{setplanmodifier.RequiresReplace()},
	}

	le := schema.BoolAttribute{
		Computed:    true,
		Description: "If true the cluster must listen HTTP port 80 because LetsEncrypt challenges there",
	}

	createdAt := r.schemaCreatedAt()

	to := timeouts.Attributes(ctx, timeouts.Opts{Create: true, Update: true, Delete: true})

	res.Schema = schema.Schema{
		Description: "Manages an AppRun dedicated cluster",
		Attributes: map[string]schema.Attribute{
			"id":                     id,
			"name":                   name,
			"service_principal_id":   spid,
			"lets_encrypt_email":     email,
			"ports":                  ports,
			"has_lets_encrypt_email": le,
			"created_at":             createdAt,
			"timeouts":               to,
		},
	}
}

func (*clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, res *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, res)
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, res *resource.CreateResponse) {
	var plan clusterResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if res.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout20min)
	defer cancel()

	created, err := r.api().Create(ctx, plan.intoCreate())

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create AppRun Dedicated cluster: %s", err))
		return
	}

	detail, err := r.api().Read(ctx, created.ClusterID)

	if err != nil {
		res.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to read created AppRun Dedicated cluster: %s", err))
		return
	}

	plan.updateState(detail)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, res *resource.ReadResponse) {
	var state clusterResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	detail, err := state.read(ctx, r, &res.Diagnostics)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read created AppRun Dedicated cluster: %s", err))
		return
	}

	state.updateState(detail)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, res *resource.UpdateResponse) {
	var plan clusterResourceModel
	res.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if res.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout20min)
	defer cancel()

	id, err := plan.clusterID()

	if err != nil {
		res.Diagnostics.AddError("Update: Invalid ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return
	}

	err = r.api().Update(ctx, id, plan.intoUpdate())

	if err != nil {
		res.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read AppRun Dedicated cluster: %s", err))
		return
	}

	detail, err := plan.read(ctx, r, &res.Diagnostics)

	if err != nil {
		res.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read created AppRun Dedicated cluster: %s", err))
		return
	}

	plan.updateState(detail)
	res.Diagnostics.Append(res.State.Set(ctx, &plan)...)
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, res *resource.DeleteResponse) {
	var state clusterResourceModel
	res.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	detail, err := state.read(ctx, r, &res.Diagnostics)

	if saclient.IsNotFoundError(err) {
		res.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated cluster: %s", err))
		return
	}

	err = r.api().Delete(ctx, detail.ClusterID)

	if err != nil {
		res.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete AppRun Dedicated cluster: %s", err))
		return
	}
}

func (r *clusterResource) api() *cluster.ClusterOp { return cluster.NewClusterOp(r.client) }

func (c *clusterResourceModel) read(ctx context.Context, res *clusterResource, d *diag.Diagnostics) (*cluster.ClusterDetail, error) {
	id, err := c.clusterID()

	if err != nil {
		d.AddError("Invalid ID", fmt.Sprintf("failed to parse cluster ID: %s", err))
		return nil, err
	}

	return res.api().Read(ctx, id)
}

func (c *clusterResourceModel) intoCreate() (ret cluster.CreateParams) {
	ret.Name = c.Name.ValueString()
	ret.ServicePrincipalID = c.ServicePrincipalID.ValueString()
	ret.LetsEncryptEmail = c.LetsEncryptEmail.ValueStringPointer()
	ret.Ports = common.MapTo(c.Ports, func(p portModel) (q v1.CreateLoadBalancerPort) {
		q.SetPort(uint16(p.Port.ValueInt32()))
		q.SetProtocol(v1.CreateLoadBalancerPortProtocol(p.Protocol.ValueString()))
		return
	})

	// `ports = []` makes no sense for us, but the API mandates empty array
	if ret.Ports == nil {
		ret.Ports = make([]v1.CreateLoadBalancerPort, 0)
	}

	return
}

func (c *clusterResourceModel) intoUpdate() (ret cluster.UpdateParams) {
	ret.ServicePrincipalID = c.ServicePrincipalID.ValueString()
	ret.LetsEncryptEmail = c.LetsEncryptEmail.ValueStringPointer()
	return
}
