// Copyright 2016-2025 terraform-provider-sakura authors
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

package script

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
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type scriptResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &scriptResource{}
	_ resource.ResourceWithConfigure   = &scriptResource{}
	_ resource.ResourceWithImportState = &scriptResource{}
)

func NewScriptResource() resource.Resource {
	return &scriptResource{}
}

func (r *scriptResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_script"
}

func (r *scriptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type scriptResourceModel struct {
	scriptBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *scriptResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":      common.SchemaResourceId("Script"),
			"name":    common.SchemaResourceName("Script"),
			"tags":    common.SchemaResourceTags("Script"),
			"icon_id": common.SchemaResourceIconID("Script"),
			"description": schema.StringAttribute{
				Computed:    true,
				Description: desc.Sprintf("The description of the Script. This will be computed from special tags within body of `content`"),
			},
			"content": schema.StringAttribute{
				Required:    true,
				Description: "The content of the Script. This must be specified as a shell script or as a cloud-config",
			},
			"class": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("shell"),
				Description: desc.Sprintf("The class of the Script. This must be one of %s", iaastypes.NoteClassStrings),
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

func (r *scriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *scriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scriptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	scriptOp := iaas.NewNoteOp(r.client)
	script, err := scriptOp.Create(ctx, &iaas.NoteCreateRequest{
		Name:    plan.Name.ValueString(),
		Tags:    common.TsetToStrings(plan.Tags),
		IconID:  common.ExpandSakuraCloudID(plan.IconID),
		Content: plan.Content.ValueString(),
		Class:   plan.Class.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud Script is failed: %s", err))
		return
	}

	plan.updateState(script)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scriptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	script := getScript(ctx, r.client, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if script == nil || resp.Diagnostics.HasError() {
		return
	}

	state.updateState(script)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan scriptResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	scriptOp := iaas.NewNoteOp(r.client)
	script, err := scriptOp.Update(ctx, common.ExpandSakuraCloudID(plan.ID), &iaas.NoteUpdateRequest{
		Name:    plan.Name.ValueString(),
		Tags:    common.TsetToStrings(plan.Tags),
		IconID:  common.ExpandSakuraCloudID(plan.IconID),
		Content: plan.Content.ValueString(),
		Class:   plan.Class.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("updating SakuraCloud Script[%s] is failed: %s", plan.ID.ValueString(), err))
		return
	}

	plan.updateState(script)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scriptResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	script := getScript(ctx, r.client, common.ExpandSakuraCloudID(state.ID), &resp.State, &resp.Diagnostics)
	if script == nil {
		return
	}

	scriptOp := iaas.NewNoteOp(r.client)
	if err := scriptOp.Delete(ctx, script.ID); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("could not delete SakuraCloud Script[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getScript(ctx context.Context, client *common.APIClient, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.Note {
	scriptOp := iaas.NewNoteOp(client)
	script, err := scriptOp.Read(ctx, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("could not read SakuraCloud Script[%s]: %s", id.String(), err))
		return nil
	}

	return script
}
