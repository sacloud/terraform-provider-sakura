// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package auto_backup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type autoBackupResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &autoBackupResource{}
	_ resource.ResourceWithConfigure   = &autoBackupResource{}
	_ resource.ResourceWithImportState = &autoBackupResource{}
)

func NewAutoBackupResource() resource.Resource {
	return &autoBackupResource{}
}

func (r *autoBackupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auto_backup"
}

func (r *autoBackupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type autoBackupResourceModel struct {
	common.SakuraBaseModel
	Zone         types.String   `tfsdk:"zone"`
	IconID       types.String   `tfsdk:"icon_id"`
	DiskID       types.String   `tfsdk:"disk_id"`
	DaysOfWeek   types.Set      `tfsdk:"days_of_week"`
	MaxBackupNum types.Int32    `tfsdk:"max_backup_num"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

func (r *autoBackupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("AutoBackup"),
			"name":        common.SchemaResourceName("AutoBackup"),
			"description": common.SchemaResourceDescription("AutoBackup"),
			"tags":        common.SchemaResourceTags("AutoBackup"),
			"zone":        common.SchemaResourceZone("AutoBackup"),
			"icon_id":     common.SchemaResourceIconID("AutoBackup"),
			"disk_id": schema.StringAttribute{
				Required:    true,
				Description: "The disk id to backed up",
				Validators: []validator.String{
					sacloudvalidator.SakuraIDValidator(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"days_of_week": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: desc.Sprintf("A list of days of week to backed up. The values in the list must be in [%s]", iaastypes.DaysOfTheWeekStrings),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf(iaastypes.DaysOfTheWeekStrings...)),
				},
			},
			"max_backup_num": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(1),
				Description: desc.Sprintf("The number backup files to keep. %s", desc.Range(1, 10)),
				Validators: []validator.Int32{
					int32validator.Between(1, 10),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manage an Auto Backup",
	}
}

func (r *autoBackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *autoBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan autoBackupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	backupOp := iaas.NewAutoBackupOp(r.client)
	created, err := backupOp.Create(ctx, zone, expandAutoBackupCreateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", "failed to create AutoBackup: "+err.Error())
		return
	}

	plan.updateState(created, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *autoBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state autoBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ab := getAutoBackup(ctx, r.client, zone, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if ab == nil {
		return
	}

	state.updateState(ab, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *autoBackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan autoBackupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ab := getAutoBackup(ctx, r.client, zone, common.ExpandSakuraCloudID(plan.ID), &resp.State, &resp.Diagnostics)
	if ab == nil {
		return
	}

	updated, err := iaas.NewAutoBackupOp(r.client).Update(ctx, zone, ab.ID, expandAutoBackupUpdateRequest(&plan, ab))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update AutoBackup[%s]: %s", ab.ID.String(), err))
		return
	}

	plan.updateState(updated, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *autoBackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state autoBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	ab := getAutoBackup(ctx, r.client, zone, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if ab == nil {
		return
	}

	if err := iaas.NewAutoBackupOp(r.client).Delete(ctx, zone, ab.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete AutoBackup[%s]: %s", ab.ID.String(), err))
		return
	}
}

func (model *autoBackupResourceModel) updateState(ab *iaas.AutoBackup, zone string) {
	model.UpdateBaseState(ab.ID.String(), ab.Name, ab.Description, ab.Tags)
	model.Zone = types.StringValue(zone)
	model.DiskID = types.StringValue(ab.DiskID.String())
	model.DaysOfWeek = common.FlattenBackupDaysOfWeek(ab.BackupSpanWeekdays)
	model.MaxBackupNum = types.Int32Value(int32(ab.MaximumNumberOfArchives))
	if ab.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(ab.IconID.String())
	}
}

func getAutoBackup(ctx context.Context, client *common.APIClient, zone string, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.AutoBackup {
	op := iaas.NewAutoBackupOp(client)
	ab, err := op.Read(ctx, zone, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read AutoBackup[%s]: %s", id.String(), err))
		return nil
	}
	return ab
}

func expandAutoBackupCreateRequest(m *autoBackupResourceModel) *iaas.AutoBackupCreateRequest {
	return &iaas.AutoBackupCreateRequest{
		Name:                    m.Name.ValueString(),
		Description:             m.Description.ValueString(),
		Tags:                    common.TsetToStrings(m.Tags),
		DiskID:                  common.ExpandSakuraCloudID(m.DiskID),
		MaximumNumberOfArchives: int(m.MaxBackupNum.ValueInt32()),
		BackupSpanWeekdays:      common.ExpandBackupDaysOfWeek(m.DaysOfWeek),
		IconID:                  common.ExpandSakuraCloudID(m.IconID),
	}
}

func expandAutoBackupUpdateRequest(m *autoBackupResourceModel, ab *iaas.AutoBackup) *iaas.AutoBackupUpdateRequest {
	return &iaas.AutoBackupUpdateRequest{
		Name:                    m.Name.ValueString(),
		Description:             m.Description.ValueString(),
		Tags:                    common.TsetToStrings(m.Tags),
		MaximumNumberOfArchives: int(m.MaxBackupNum.ValueInt32()),
		BackupSpanWeekdays:      common.ExpandBackupDaysOfWeek(m.DaysOfWeek),
		IconID:                  common.ExpandSakuraCloudID(m.IconID),
		SettingsHash:            ab.SettingsHash,
	}
}
