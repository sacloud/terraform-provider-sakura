// Copyright 2016-2025 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sakura

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/accessor"
	"github.com/sacloud/iaas-api-go/helper/power"
	"github.com/sacloud/iaas-api-go/helper/query"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/iaas-service-go/setup"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

type nfsResource struct {
	client *APIClient
}

var (
	_ resource.Resource                = &nfsResource{}
	_ resource.ResourceWithConfigure   = &nfsResource{}
	_ resource.ResourceWithImportState = &nfsResource{}
)

func NewNFSResource() resource.Resource {
	return &nfsResource{}
}

func (r *nfsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nfs"
}

func (r *nfsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type nfsResourceModel struct {
	sakuraNFSBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *nfsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaResourceId("NFS"),
			"name":        schemaResourceName("NFS"),
			"plan":        schemaResourcePlan("NFS", "hdd", iaastypes.NFSPlanStrings),
			"size":        schemaResourceSize("NFS", 100),
			"icon_id":     schemaResourceIconID("NFS"),
			"description": schemaResourceDescription("NFS"),
			"tags":        schemaResourceTags("NFS"),
			"zone":        schemaResourceZone("NFS"),
		},
		Blocks: map[string]schema.Block{
			"network_interface": schema.ListNestedBlock{ // データの互換性のためにListにしているが、SingleNestedBlockが望ましい
				Description: "The network interface of the NFS.",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"switch_id": schemaResourceSwitchID("NFS"),
						"ip_address": schema.StringAttribute{
							Required:    true,
							Description: "The IP address to assign to the NFS",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"netmask": schema.Int32Attribute{
							Required:    true,
							Description: desc.Sprintf("The bit length of the subnet to assign to the NFS. %s", desc.Range(8, 29)),
							Validators: []validator.Int32{
								int32validator.Between(8, 29),
							},
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
						},
						"gateway": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The IP address of the gateway used by NFS",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplaceIfConfigured(),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *nfsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *nfsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nfsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutCreate(ctx, plan.Timeouts, 24*time.Hour)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	nfsOp := iaas.NewNFSOp(r.client)
	planID, err := expandNFSDiskPlanID(ctx, r.client, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", err.Error())
		return
	}

	builder := &setup.RetryableSetup{
		Create: func(ctx context.Context, zone string) (accessor.ID, error) {
			return nfsOp.Create(ctx, zone, expandNFSCreateRequest(&plan, planID))
		},
		Delete: func(ctx context.Context, zone string, id iaastypes.ID) error {
			return nfsOp.Delete(ctx, zone, id)
		},
		Read: func(ctx context.Context, zone string, id iaastypes.ID) (interface{}, error) {
			return nfsOp.Read(ctx, zone, id)
		},
		IsWaitForCopy: true,
		IsWaitForUp:   true,
		Options: &setup.Options{
			RetryCount: 3,
		},
	}

	res, err := builder.Setup(ctx, zone)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud NFS is failed: %s", err))
		return
	}

	nfs, ok := res.(*iaas.NFS)
	if !ok {
		resp.Diagnostics.AddError("Create Error", "creating SakuraCloud NFS is failed: created resource is not *iaas.NFS")
		return
	}

	updateResourceByReadWithZone(ctx, r, &resp.State, &resp.Diagnostics, nfs.ID.String(), zone)
}

func (r *nfsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nfsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	nfsOp := iaas.NewNFSOp(r.client)
	nfs, err := nfsOp.Read(ctx, zone, expandSakuraCloudID(state.ID))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", err.Error())
		return
	}

	if rmResource, err := state.updateState(ctx, r.client, nfs, zone); err != nil {
		if rmResource {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not update state for SakuraCloud NFS resource: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *nfsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan nfsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutUpdate(ctx, plan.Timeouts, 24*time.Hour)
	defer cancel()

	zone := getZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	nfsOp := iaas.NewNFSOp(r.client)
	nfs, err := nfsOp.Read(ctx, zone, expandSakuraCloudID(plan.ID))
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("could not read SakuraCloud NFS[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	_, err = nfsOp.Update(ctx, zone, nfs.ID, expandNFSUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("updating SakuraCloud NFS[%s] is failed: %s", plan.ID.ValueString(), err))
		return
	}

	updateResourceByReadWithZone(ctx, r, &resp.State, &resp.Diagnostics, nfs.ID.String(), zone)
}

func (r *nfsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nfsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutDelete(ctx, state.Timeouts, timeout20min)
	defer cancel()

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	nfsOp := iaas.NewNFSOp(r.client)
	nfs, err := nfsOp.Read(ctx, zone, sakuraCloudID(state.ID.ValueString()))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("could not read SakuraCloud NFS[%s]: %s", state.ID.ValueString(), err))
		return
	}

	if err := power.ShutdownNFS(ctx, nfsOp, zone, nfs.ID, true); err != nil {
		resp.Diagnostics.AddError("Delete Error", err.Error())
		return
	}

	if err := nfsOp.Delete(ctx, zone, nfs.ID); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("deleting SakuraCloud NFS[%s] is failed: %s", state.ID.ValueString(), err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func expandNFSDiskPlanID(ctx context.Context, client *APIClient, d *nfsResourceModel) (iaastypes.ID, error) {
	var planID iaastypes.ID
	planName := d.Plan.ValueString()
	planID, ok := iaastypes.NFSPlanIDMap[planName]
	if !ok {
		return iaastypes.ID(0), fmt.Errorf("plan is not found: %s", planName)
	}
	size := d.Size.ValueInt64()

	return query.FindNFSPlanID(ctx, iaas.NewNoteOp(client), planID, iaastypes.ENFSSize(size))
}

func expandNFSCreateRequest(d *nfsResourceModel, planID iaastypes.ID) *iaas.NFSCreateRequest {
	nic := d.NetworkInterface[0]
	return &iaas.NFSCreateRequest{
		SwitchID:       expandSakuraCloudID(nic.SwitchID),
		PlanID:         planID,
		IPAddresses:    []string{nic.IPAddress.ValueString()},
		NetworkMaskLen: int(nic.Netmask.ValueInt32()),
		DefaultRoute:   nic.Gateway.ValueString(),
		Name:           d.Name.ValueString(),
		Description:    d.Description.ValueString(),
		Tags:           tsetToStrings(d.Tags),
		IconID:         expandSakuraCloudID(d.IconID),
	}
}

func expandNFSUpdateRequest(d *nfsResourceModel) *iaas.NFSUpdateRequest {
	return &iaas.NFSUpdateRequest{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Tags:        tsetToStrings(d.Tags),
		IconID:      expandSakuraCloudID(d.IconID),
	}
}
