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

package note

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"

	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

type noteResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &noteResource{}
	_ resource.ResourceWithConfigure   = &noteResource{}
	_ resource.ResourceWithImportState = &noteResource{}
)

func NewNoteResource() resource.Resource {
	return &noteResource{}
}

func (r *noteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_note"
}

func (r *noteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type noteResourceModel struct {
	noteBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *noteResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":      common.SchemaResourceId("Note"),
			"name":    common.SchemaResourceName("Note"),
			"tags":    common.SchemaResourceTags("Note"),
			"icon_id": common.SchemaResourceIconID("Note"),
			"description": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The description of the Note. This will be computed from special tags within body of `content`"),
			},
			"content": schema.StringAttribute{
				Required:    true,
				Description: "The content of the Note. This must be specified as a shell script or as a cloud-config",
			},
			"class": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("shell"),
				Description: desc.Sprintf("The class of the Note. This must be one of %s", iaastypes.NoteClassStrings),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.NoteClassStrings...),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *noteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *noteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan noteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	noteOp := iaas.NewNoteOp(r.client)
	note, err := noteOp.Create(ctx, &iaas.NoteCreateRequest{
		Name:    plan.Name.ValueString(),
		Tags:    common.TsetToStrings(plan.Tags),
		IconID:  common.ExpandSakuraCloudID(plan.IconID),
		Content: plan.Content.ValueString(),
		Class:   plan.Class.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud Note is failed: %s", err))
		return
	}

	plan.updateState(note)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *noteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state noteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	note := getNote(ctx, r.client, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if note == nil || resp.Diagnostics.HasError() {
		return
	}

	state.updateState(note)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *noteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan noteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	noteOp := iaas.NewNoteOp(r.client)
	note, err := noteOp.Update(ctx, common.ExpandSakuraCloudID(plan.ID), &iaas.NoteUpdateRequest{
		Name:    plan.Name.ValueString(),
		Tags:    common.TsetToStrings(plan.Tags),
		IconID:  common.ExpandSakuraCloudID(plan.IconID),
		Content: plan.Content.ValueString(),
		Class:   plan.Class.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("updating SakuraCloud Note[%s] is failed: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(note)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *noteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state noteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	noteOp := iaas.NewNoteOp(r.client)
	note := getNote(ctx, r.client, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if note == nil {
		return
	}

	if err := noteOp.Delete(ctx, note.ID); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("could not delete SakuraCloud Note[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getNote(ctx context.Context, client *common.APIClient, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.Note {
	noteOp := iaas.NewNoteOp(client)
	note, err := noteOp.Read(ctx, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("could not read SakuraCloud Note[%s]: %s", id.String(), err))
		return nil
	}

	return note
}
