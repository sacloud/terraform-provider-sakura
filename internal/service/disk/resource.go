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

package disk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/accessor"
	"github.com/sacloud/iaas-api-go/helper/cleanup"
	"github.com/sacloud/iaas-api-go/helper/power"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/iaas-service-go/setup"
	"github.com/sacloud/packages-go/size"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/validators"
)

type diskResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &diskResource{}
	_ resource.ResourceWithConfigure   = &diskResource{}
	_ resource.ResourceWithImportState = &diskResource{}
)

func NewDiskResource() resource.Resource {
	return &diskResource{}
}

func (r *diskResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_disk"
}

func (r *diskResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type diskResourceModel struct {
	diskBaseModel
	DistantFrom types.Set      `tfsdk:"distant_from"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (r *diskResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Disk"),
			"name":        common.SchemaResourceName("Disk"),
			"description": common.SchemaResourceDescription("Disk"),
			"tags":        common.SchemaResourceTags("Disk"),
			"zone":        common.SchemaResourceZone("Disk"),
			"icon_id":     common.SchemaResourceIconID("Disk"),
			"size":        common.SchemaResourceSize("Disk", 20),
			"plan":        common.SchemaResourcePlan("Disk", iaastypes.DiskPlanNameMap[iaastypes.DiskPlans.SSD], iaastypes.DiskPlanStrings),
			"server_id":   common.SchemaResourceServerID("Disk"),
			"connector": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(iaastypes.DiskConnections.VirtIO.String()),
				Description: desc.Sprintf("The name of the disk connector. This must be one of [%s]", iaastypes.DiskConnectionStrings),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.DiskConnectionStrings...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"source_archive_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The id of the source archive. %s", desc.Conflicts("source_disk_id")),
				Validators: []validator.String{
					validators.SakuraIDValidator(),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("source_disk_id")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"source_disk_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The id of the source disk. %s", desc.Conflicts("source_archive_id")),
				Validators: []validator.String{
					validators.SakuraIDValidator(),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("source_archive_id")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"encryption_algorithm": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(iaastypes.DiskEncryptionAlgorithms.None.String()),
				Description: desc.Sprintf("The disk encryption algorithm. This must be one of [%s]", iaastypes.DiskEncryptionAlgorithmStrings),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.DiskEncryptionAlgorithmStrings...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"distant_from": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of disk id. The disk will be located to different storage from these disks",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.SakuraIDValidator()),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *diskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *diskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan diskResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout24hour)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diskOp := iaas.NewDiskOp(r.client)
	diskBuilder := &setup.RetryableSetup{
		IsWaitForCopy: true,
		Create: func(ctx context.Context, zone string) (accessor.ID, error) {
			return diskOp.Create(ctx, zone, expandDiskCreateRequest(&plan), common.ExpandSakuraCloudIDs(plan.DistantFrom))
		},
		Read: func(ctx context.Context, zone string, id iaastypes.ID) (interface{}, error) {
			return diskOp.Read(ctx, zone, id)
		},
		Delete: func(ctx context.Context, zone string, id iaastypes.ID) error {
			return diskOp.Delete(ctx, zone, id)
		},
		Options: &setup.Options{
			RetryCount: 3,
		},
	}

	res, err := diskBuilder.Setup(ctx, zone)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud Disk is failed: %s", err))
		return
	}

	disk, ok := res.(*iaas.Disk)
	if !ok {
		resp.Diagnostics.AddError("Create Error", "creating SakuraCloud Disk is failed: created resource is not a *iaas.Disk")
		return
	}

	plan.updateState(disk, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *diskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state diskResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	disk := getDisk(ctx, r.client, common.ExpandSakuraCloudID(state.ID), zone, &resp.State, &resp.Diagnostics)
	if disk == nil {
		return
	}

	state.updateState(disk, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *diskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan diskResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout24hour)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diskOp := iaas.NewDiskOp(r.client)
	_, err := diskOp.Update(ctx, zone, common.ExpandSakuraCloudID(plan.ID), expandDiskUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("updating SakuraCloud Disk[%s] is failed: %s", plan.ID.ValueString(), err))
		return
	}

	disk := getDisk(ctx, r.client, common.ExpandSakuraCloudID(plan.ID), zone, &resp.State, &resp.Diagnostics)
	if disk == nil {
		return
	}

	plan.updateState(disk, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *diskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state diskResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout20min)
	defer cancel()

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	diskOp := iaas.NewDiskOp(r.client)
	disk := getDisk(ctx, r.client, common.ExpandSakuraCloudID(state.ID), zone, &resp.State, &resp.Diagnostics)
	if disk == nil {
		return
	}

	serverID := disk.GetServerID()
	if serverID != 0 {
		serverOp := iaas.NewServerOp(r.client)
		server, err := serverOp.Read(ctx, zone, serverID)
		if err != nil {
			resp.Diagnostics.AddError("Delete Error",
				fmt.Sprintf("could not read SakuraCloud Server[%s] of Disk[%s]: %s", serverID.String(), disk.ID.String(), err))
			return
		}

		if server.InstanceStatus.IsUp() {
			if err := power.ShutdownServer(ctx, serverOp, zone, server.ID, true); err != nil {
				resp.Diagnostics.AddError("Delete Error",
					fmt.Sprintf("stopping SakuraCloud Server[%s] of Disk[%s] is failed: %s", server.ID.String(), disk.ID.String(), err))
				return
			}
		}

		if err := diskOp.DisconnectFromServer(ctx, zone, disk.ID); err != nil {
			resp.Diagnostics.AddError("Delete Error",
				fmt.Sprintf("disconnect from SakuraCloud Server[%s] of Disk[%s] is failed: %s", server.ID.String(), disk.ID.String(), err))
			return
		}
	}

	if err := cleanup.DeleteDisk(ctx, r.client, zone, disk.ID, r.client.CheckReferencedOption()); err != nil {
		resp.Diagnostics.AddError("Delete Error",
			fmt.Sprintf("deleting SakuraCloud Disk[%s] is failed: %s", disk.ID.String(), err))
		return
	}
}

func getDisk(ctx context.Context, client *common.APIClient, id iaastypes.ID, zone string, state *tfsdk.State, diags *diag.Diagnostics) *iaas.Disk {
	diskOp := iaas.NewDiskOp(client)
	disk, err := diskOp.Read(ctx, zone, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Get Disk Error", fmt.Sprintf("could not read SakuraCloud Disk[%s]: %s", id.String(), err))
		return nil
	}

	return disk
}

func expandDiskCreateRequest(d *diskResourceModel) *iaas.DiskCreateRequest {
	return &iaas.DiskCreateRequest{
		DiskPlanID:          iaastypes.DiskPlanIDMap[d.Plan.ValueString()],
		Connection:          iaastypes.EDiskConnection(d.Connector.ValueString()),
		ServerID:            common.ExpandSakuraCloudID(d.ServerID),
		SourceDiskID:        common.ExpandSakuraCloudID(d.SourceDiskID),
		SourceArchiveID:     common.ExpandSakuraCloudID(d.SourceArchiveID),
		SizeMB:              int(d.Size.ValueInt64()) * size.GiB,
		Name:                d.Name.ValueString(),
		Description:         d.Description.ValueString(),
		Tags:                common.TsetToStrings(d.Tags),
		IconID:              common.ExpandSakuraCloudID(d.IconID),
		EncryptionAlgorithm: iaastypes.EDiskEncryptionAlgorithm(d.EncryptionAlgorithm.ValueString()),
	}
}

func expandDiskUpdateRequest(d *diskResourceModel) *iaas.DiskUpdateRequest {
	return &iaas.DiskUpdateRequest{
		Connection:  iaastypes.EDiskConnection(d.Connector.ValueString()),
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Tags:        common.TsetToStrings(d.Tags),
		IconID:      common.ExpandSakuraCloudID(d.IconID),
	}
}
