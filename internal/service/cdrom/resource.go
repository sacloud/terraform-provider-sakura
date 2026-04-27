// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package cdrom

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	validator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/cleanup"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	cdromsvc "github.com/sacloud/iaas-service-go/cdrom"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	"github.com/sacloud/terraform-provider-sakura/internal/ftps"
)

var cdromValidSizes = []int32{5, 10, 20}

type cdromResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &cdromResource{}
	_ resource.ResourceWithConfigure   = &cdromResource{}
	_ resource.ResourceWithImportState = &cdromResource{}
)

func NewCDROMResource() resource.Resource {
	return &cdromResource{}
}

func (r *cdromResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdrom"
}

func (r *cdromResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type cdromResourceModel struct {
	common.SakuraBaseModel
	Zone         types.String   `tfsdk:"zone"`
	Size         types.Int32    `tfsdk:"size"`
	IconID       types.String   `tfsdk:"icon_id"`
	ISOImageFile types.String   `tfsdk:"iso_image_file"`
	Hash         types.String   `tfsdk:"hash"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

func (r *cdromResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("CD-ROM"),
			"name":        common.SchemaResourceName("CD-ROM"),
			"icon_id":     common.SchemaResourceIconID("CD-ROM"),
			"description": common.SchemaResourceDescription("CD-ROM"),
			"tags":        common.SchemaResourceTags("CD-ROM"),
			"zone":        common.SchemaResourceZone("CD-ROM"),
			"size": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int32default.StaticInt32(5),
				Description: desc.Sprintf("The size of CD-ROM in GiB. This must be one of [%s]", cdromValidSizes),
				Validators: []validator.Int32{
					int32validator.OneOf(cdromValidSizes...),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"iso_image_file": schema.StringAttribute{
				Required:    true,
				Description: "The file path to upload to as the CD-ROM.",
			},
			"hash": schema.StringAttribute{
				Computed:    true,
				Description: "The md5 checksum calculated from the uploaded ISO file",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages a CD-ROM.",
	}
}

func (r *cdromResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *cdromResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cdromResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, 24*time.Hour)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	sourcePath, err := common.ExpandHomeDir(plan.ISOImageFile.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create: File Error", err.Error())
		return
	}

	svc := cdromsvc.New(r.client)
	cdrom, err := svc.CreateWithContext(ctx, &cdromsvc.CreateRequest{
		Zone:        zone,
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Tags:        common.TsetToStrings(plan.Tags),
		IconID:      common.ExpandSakuraCloudID(plan.IconID),
		SizeGB:      int(plan.Size.ValueInt32()),
		SourcePath:  sourcePath,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create CD-ROM: %s", err))
		return
	}

	plan.updateState(cdrom, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cdromResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cdromResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := common.GetZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	cdrom := getCDROM(ctx, r.client, common.ExpandSakuraCloudID(state.ID), zone, &resp.State, &resp.Diagnostics)
	if cdrom == nil || resp.Diagnostics.HasError() {
		return
	}

	state.updateState(cdrom, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cdromResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state cdromResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout24hour)
	defer cancel()

	zone := common.GetZone(plan.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	id := common.ExpandSakuraCloudID(plan.ID)
	svc := cdromsvc.New(r.client)

	name := plan.Name.ValueString()
	description := plan.Description.ValueString()
	tags := iaastypes.Tags(common.TsetToStrings(plan.Tags))
	iconID := common.ExpandSakuraCloudID(plan.IconID)
	if _, err := svc.UpdateWithContext(ctx, &cdromsvc.UpdateRequest{
		Zone:        zone,
		ID:          id,
		Name:        &name,
		Description: &description,
		Tags:        &tags,
		IconID:      &iconID,
	}); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update CD-ROM[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	newHash := expandCDROMHash(&plan)
	if newHash != state.Hash.ValueString() {
		sourcePath, err := common.ExpandHomeDir(plan.ISOImageFile.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Update: File Error", err.Error())
			return
		}

		ejectedServerIDs, err := ejectCDROMFromAllServers(ctx, r.client, zone, id)
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("could not eject CD-ROM[%s] from Server: %s", id, err))
			return
		}

		// ejectしたサーバへは成否にかかわらず再挿入を試みる。
		// 途中で失敗すると利用者側のサーバから CDROM が外れたまま残ってしまうため、
		// context が切れているケースに備えて独立したタイムアウト付きの context を用いる。
		defer func() {
			insertCtx, cancel := context.WithTimeout(context.Background(), common.Timeout20min)
			defer cancel()
			if err := insertCDROMToAllServers(insertCtx, r.client, zone, id, ejectedServerIDs); err != nil {
				resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("could not insert CD-ROM[%s] to Server: %s", id, err))
			}
		}()

		cdromOp := iaas.NewCDROMOp(r.client)
		ftpServer, err := cdromOp.OpenFTP(ctx, zone, id, &iaas.OpenFTPRequest{ChangePassword: false})
		if err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("opening FTP connection to CD-ROM[%s] is failed: %s", id, err))
			return
		}

		if err := ftps.UploadFile(ctx, ftpServer.User, ftpServer.Password, ftpServer.HostName, sourcePath); err != nil {
			resp.Diagnostics.AddError("Update: Upload Error", fmt.Sprintf("failed to upload CD-ROM[%s]: %s", id, err))
			return
		}

		if err := cdromOp.CloseFTP(ctx, zone, id); err != nil {
			resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("closing FTP connection is failed: %s", err))
			return
		}
	}

	cdrom := getCDROM(ctx, r.client, id, zone, &resp.State, &resp.Diagnostics)
	if cdrom == nil || resp.Diagnostics.HasError() {
		return
	}

	plan.updateState(cdrom, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cdromResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cdromResourceModel
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

	cdrom := getCDROM(ctx, r.client, common.ExpandSakuraCloudID(state.ID), zone, &resp.State, &resp.Diagnostics)
	if cdrom == nil || resp.Diagnostics.HasError() {
		return
	}

	// サーバ破棄直後は API 側で一時的に参照が残る場合があるため、参照解放を待ってから削除する
	if err := cleanup.DeleteCDROM(ctx, r.client, zone, cdrom.ID, r.client.CheckReferencedOption()); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete CD-ROM[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func (model *cdromResourceModel) updateState(cdrom *iaas.CDROM, zone string) {
	model.UpdateBaseState(cdrom.ID.String(), cdrom.Name, cdrom.Description, cdrom.Tags)
	model.Size = types.Int32Value(int32(cdrom.GetSizeGB()))
	model.Zone = types.StringValue(zone)
	model.Hash = types.StringValue(expandCDROMHash(model))
	if cdrom.IconID.IsEmpty() {
		model.IconID = types.StringNull()
	} else {
		model.IconID = types.StringValue(cdrom.IconID.String())
	}
}

func getCDROM(ctx context.Context, client *common.APIClient, id iaastypes.ID, zone string, state *tfsdk.State, diags *diag.Diagnostics) *iaas.CDROM {
	cdromOp := iaas.NewCDROMOp(client)
	cdrom, err := cdromOp.Read(ctx, zone, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read CD-ROM[%s]: %s", id, err))
		return nil
	}

	return cdrom
}

func expandCDROMHash(d *cdromResourceModel) string {
	source := d.ISOImageFile.ValueString()
	if source == "" {
		return ""
	}

	path, err := common.ExpandHomeDir(source)
	if err != nil {
		return ""
	}

	// NOTE 本来はAPIにてmd5ハッシュを取得できるのが望ましい。現状ではここでファイルを読んで算出する。
	hash, err := common.Md5CheckSumFromFile(path)
	if err != nil {
		return ""
	}
	return hash
}

func ejectCDROMFromAllServers(ctx context.Context, client *common.APIClient, zone string, cdromID iaastypes.ID) ([]iaastypes.ID, error) {
	serverOp := iaas.NewServerOp(client)
	searched, err := serverOp.Find(ctx, zone, &iaas.FindCondition{})
	if err != nil {
		return nil, err
	}
	var ejectedIDs []iaastypes.ID
	for _, server := range searched.Servers {
		if server.CDROMID == cdromID {
			if err := serverOp.EjectCDROM(ctx, zone, server.ID, &iaas.EjectCDROMRequest{ID: cdromID}); err != nil {
				return nil, err
			}
			ejectedIDs = append(ejectedIDs, server.ID)
		}
	}
	return ejectedIDs, nil
}

func insertCDROMToAllServers(ctx context.Context, client *common.APIClient, zone string, cdromID iaastypes.ID, serverIDs []iaastypes.ID) error {
	serverOp := iaas.NewServerOp(client)
	for _, id := range serverIDs {
		if err := serverOp.InsertCDROM(ctx, zone, id, &iaas.InsertCDROMRequest{ID: cdromID}); err != nil {
			return err
		}
	}
	return nil
}
