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
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	archiveUtil "github.com/sacloud/iaas-service-go/archive/builder"

	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

type archiveResource struct {
	client *APIClient
}

var (
	_ resource.Resource                = &archiveResource{}
	_ resource.ResourceWithConfigure   = &archiveResource{}
	_ resource.ResourceWithImportState = &archiveResource{}
)

func NewArchiveResource() resource.Resource {
	return &archiveResource{}
}

func (r *archiveResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_archive"
}

func (r *archiveResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

// TODO: model.goに切り出してdata sourceと共通化する
type archiveResourceModel struct {
	sakuraBaseModel
	Zone              types.String   `tfsdk:"zone"`
	Size              types.Int32    `tfsdk:"size"`
	Hash              types.String   `tfsdk:"hash"`
	IconID            types.String   `tfsdk:"icon_id"`
	ArchiveFile       types.String   `tfsdk:"archive_file"`
	SourceDiskID      types.String   `tfsdk:"source_disk_id"`
	SourceSharedKey   types.String   `tfsdk:"source_shared_key"`
	SourceArchiveID   types.String   `tfsdk:"source_archive_id"`
	SourceArchiveZone types.String   `tfsdk:"source_archive_zone"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func (r *archiveResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	sizePath := path.MatchRelative().AtParent().AtName("size")
	sourceDiskIdPath := path.MatchRelative().AtParent().AtName("source_disk_id")
	sourceSharedKeyPath := path.MatchRelative().AtParent().AtName("source_shared_key")
	sourceArchiveIdPath := path.MatchRelative().AtParent().AtName("source_archive_id")
	sourceArchiveZonePath := path.MatchRelative().AtParent().AtName("source_archive_zone")

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaResourceId("Archive"),
			"name":        schemaResourceName("Archive"),
			"icon_id":     schemaResourceIconID("Archive"),
			"description": schemaResourceDescription("Archive"),
			"tags":        schemaResourceTags("Archive"),
			"zone":        schemaResourceZone("Archive"),
			"size": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The size of archihve in GiB. This must be one of [%s]", iaastypes.ArchiveSizes),
				Validators: []validator.Int32{
					int32validator.OneOf(mapTo(iaastypes.ArchiveSizes, intToInt32)...),
					int32validator.ConflictsWith(sourceDiskIdPath, sourceSharedKeyPath, sourceArchiveIdPath, sourceArchiveZonePath),
				},
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"archive_file": schema.StringAttribute{
				Optional:    true,
				Description: "The file path to upload to the SakuraCloud Archive.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"hash": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The md5 checksum calculated from the base64 encoded file body",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"source_archive_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The id of the source archive. %s", desc.Conflicts("source_disk_id")),
				Validators: []validator.String{
					sakuraIDValidator(),
					stringvalidator.ConflictsWith(sizePath, sourceDiskIdPath, sourceSharedKeyPath),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"source_archive_zone": schema.StringAttribute{
				Optional:    true,
				Description: "The share key of source shared archive",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(sizePath, sourceDiskIdPath, sourceSharedKeyPath),
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
					sakuraIDValidator(),
					stringvalidator.ConflictsWith(sizePath, sourceArchiveIdPath, sourceArchiveZonePath, sourceSharedKeyPath),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"source_shared_key": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "The share key of source shared archive",
				Validators: []validator.String{
					stringFuncValidator(func(v string) error {
						key := iaastypes.ArchiveShareKey(v)
						if !key.ValidFormat() {
							return fmt.Errorf("%q must be formatted in '<ZONE>:<ID>:<TOKEN>'", key)
						}
						return nil
					}),
					stringvalidator.ConflictsWith(sizePath, sourceArchiveIdPath, sourceArchiveZonePath, sourceDiskIdPath),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *archiveResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *archiveResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan archiveResourceModel
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

	builder, cleanup, err := expandArchiveBuilder(&plan, zone, r.client)
	if err != nil {
		resp.Diagnostics.AddError("Archive builder error", err.Error())
		return
	}
	if cleanup != nil {
		defer cleanup()
	}

	archive, err := builder.Build(ctx, zone)
	if err != nil {
		resp.Diagnostics.AddError("Archive Create Error", fmt.Sprintf("creating SakuraCloud Archive is failed: %s", err))
		return
	}
	// Read(API)で取得できないoptionalな値をStateに設定するのは出来ないため、ここでresp.State.Setを呼び出しておく
	plan.updateState(r.client, archive, zone)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

	updateResourceByReadWithZone(ctx, r, &resp.State, &resp.Diagnostics, archive.ID.String(), zone)
}

func (r *archiveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state archiveResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := getZone(state.Zone, r.client, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	archiveOp := iaas.NewArchiveOp(r.client)
	archive, err := archiveOp.Read(ctx, zone, sakuraCloudID(state.ID.ValueString()))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not read SakuraCloud Archive[%s]: %s", state.ID.ValueString(), err))
		return
	}

	state.updateState(r.client, archive, zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *archiveResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan archiveResourceModel
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

	archiveOp := iaas.NewArchiveOp(r.client)
	archive, err := archiveOp.Read(ctx, zone, sakuraCloudID(plan.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("reading SakuraCloud Archive[%s] is failed: %s", plan.ID.ValueString(), err))
		return
	}

	if _, err = archiveOp.Update(ctx, zone, archive.ID, expandArchiveUpdateRequest(&plan)); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("updating SakuraCloud Archive[%s] is failed: %s", plan.ID.ValueString(), err))
		return
	}

	updateResourceByReadWithZone(ctx, r, &resp.State, &resp.Diagnostics, plan.ID.ValueString(), zone)
}

func (r *archiveResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state archiveResourceModel
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

	archiveOp := iaas.NewArchiveOp(r.client)
	archive, err := archiveOp.Read(ctx, zone, sakuraCloudID(state.ID.ValueString()))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("could not read SakuraCloud Archive[%s]: %s", state.ID.ValueString(), err))
		return
	}

	if err := archiveOp.Delete(ctx, zone, archive.ID); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("deleting SakuraCloud Archive[%s] is failed: %s", state.ID.ValueString(), err))
		return
	}

	resp.State.RemoveResource(ctx)
}

func (model *archiveResourceModel) updateState(client *APIClient, archive *iaas.Archive, zone string) {
	model.updateBaseState(archive.ID.String(), archive.Name, archive.Description, archive.Tags)
	model.IconID = types.StringValue(archive.IconID.String())
	model.Size = types.Int32Value(int32(archive.GetSizeGB()))
	model.Zone = types.StringValue(zone)
	model.Hash = types.StringValue(expandArchiveHash(model))
	model.SourceArchiveID = types.StringValue(model.SourceArchiveID.ValueString())
	model.SourceDiskID = types.StringValue(model.SourceDiskID.ValueString())
	model.SourceSharedKey = types.StringValue(model.SourceSharedKey.ValueString())
}

func expandArchiveBuilder(d *archiveResourceModel, zone string, client *APIClient) (archiveUtil.Builder, func(), error) {
	var reader io.ReadCloser
	source := d.ArchiveFile.ValueString()
	if source != "" {
		sourcePath, err := expandHomeDir(source)
		if err != nil {
			return nil, nil, err
		}
		f, err := os.Open(filepath.Clean(sourcePath))
		if err != nil {
			return nil, nil, err
		}
		reader = f
	}

	sourceArchiveZone := d.SourceArchiveZone.ValueString()
	if sourceArchiveZone != "" {
		if err := StringInSlice(client.zones, "source_archive_zone", sourceArchiveZone, false); err != nil {
			return nil, nil, err
		}
		if zone == sourceArchiveZone {
			sourceArchiveZone = ""
		}
	}
	sizeGB := d.Size.ValueInt32()
	if sizeGB == 0 {
		sizeGB = 20
	}

	// Note: APIとしてはディスクやアーカイブをソースとした場合Sizeの指定はできないが、
	//       archiveUtil.Director側でAPIに渡すパラメータを制御しているためここでは常に渡して問題ない
	director := &archiveUtil.Director{
		Name:              d.Name.ValueString(),
		Description:       d.Description.ValueString(),
		Tags:              tsetToStrings(d.Tags),
		IconID:            expandSakuraCloudID(d.IconID),
		SizeGB:            int(sizeGB),
		SourceReader:      reader,
		SourceDiskID:      expandSakuraCloudID(d.SourceDiskID),
		SourceArchiveID:   expandSakuraCloudID(d.SourceArchiveID),
		SourceArchiveZone: sourceArchiveZone,
		SourceSharedKey:   iaastypes.ArchiveShareKey(d.SourceSharedKey.ValueString()),
		Client:            archiveUtil.NewAPIClient(client),
	}
	return director.Builder(), func() {
		if reader != nil {
			reader.Close() //nolint
		}
	}, nil
}

func expandArchiveHash(d *archiveResourceModel) string {
	source := d.ArchiveFile.ValueString()
	if source == "" {
		return ""
	}

	path, err := expandHomeDir(source)
	if err != nil {
		return ""
	}

	// NOTE 本来はAPIにてmd5ハッシュを取得できるのが望ましい。現状ではここでファイルを読んで算出する。
	hash, err := md5CheckSumFromFile(path)
	if err != nil {
		return ""
	}
	return hash
}

func expandArchiveUpdateRequest(d *archiveResourceModel) *iaas.ArchiveUpdateRequest {
	return &iaas.ArchiveUpdateRequest{
		Name:        d.Name.ValueString(),
		Description: d.Description.ValueString(),
		Tags:        tsetToStrings(d.Tags),
		IconID:      expandSakuraCloudID(d.IconID),
	}
}
