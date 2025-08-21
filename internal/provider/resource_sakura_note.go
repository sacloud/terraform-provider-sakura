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

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	iaas "github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

type noteResource struct {
	client *APIClient
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
	apiclient := getApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

type noteResourceModel struct {
	sakuraNoteBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *noteResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":      schemaResourceId("Note"),
			"name":    schemaDataSourceName("Note"),
			"tags":    schemaResourceTags("Note"),
			"icon_id": schemaResourceIconID("Note"),
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
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
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

	ctx, cancel := setupTimeoutCreate(ctx, plan.Timeouts, timeout5min)
	defer cancel()

	noteOp := iaas.NewNoteOp(r.client)
	note, err := noteOp.Create(ctx, &iaas.NoteCreateRequest{
		Name:    plan.Name.ValueString(),
		Tags:    tsetToStrings(plan.Tags),
		IconID:  expandSakuraCloudID(plan.IconID),
		Content: plan.Content.ValueString(),
		Class:   plan.Class.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud Note is failed: %s", err))
		return
	}

	updateResourceByRead(ctx, r, &resp.State, &resp.Diagnostics, note.ID.String())
}

func (r *noteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state noteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	noteOp := iaas.NewNoteOp(r.client)
	note, err := noteOp.Read(ctx, expandSakuraCloudID(state.ID))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not read SakuraCloud Note[%s]: %s", state.ID.ValueString(), err))
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

	ctx, cancel := setupTimeoutUpdate(ctx, plan.Timeouts, timeout5min)
	defer cancel()

	noteOp := iaas.NewNoteOp(r.client)
	note, err := noteOp.Update(ctx, expandSakuraCloudID(plan.ID), &iaas.NoteUpdateRequest{
		Name:    plan.Name.ValueString(),
		Tags:    tsetToStrings(plan.Tags),
		IconID:  expandSakuraCloudID(plan.IconID),
		Content: plan.Content.ValueString(),
		Class:   plan.Class.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("updating SakuraCloud Note[%s] is failed: %s", plan.ID.ValueString(), err))
		return
	}

	updateResourceByRead(ctx, r, &resp.State, &resp.Diagnostics, note.ID.String())
}

func (r *noteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state noteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setupTimeoutDelete(ctx, state.Timeouts, timeout5min)
	defer cancel()

	noteOp := iaas.NewNoteOp(r.client)
	note, err := noteOp.Read(ctx, expandSakuraCloudID(state.ID))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("could not read SakuraCloud Note[%s]: %s", state.ID.ValueString(), err))
		return
	}

	if err := noteOp.Delete(ctx, note.ID); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("could not delete SakuraCloud Note[%s]: %s", state.ID.ValueString(), err))
		return
	}

	resp.State.RemoveResource(ctx)
}
