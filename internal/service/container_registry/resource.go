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

package container_registry

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	registryBuilder "github.com/sacloud/iaas-service-go/containerregistry/builder"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

type containerRegistryResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &containerRegistryResource{}
	_ resource.ResourceWithConfigure   = &containerRegistryResource{}
	_ resource.ResourceWithImportState = &containerRegistryResource{}
)

func NewContainerRegistryResource() resource.Resource {
	return &containerRegistryResource{}
}

func (r *containerRegistryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_registry"
}

func (r *containerRegistryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type containerRegistryResourceModel struct {
	sakuraContainerRegistryBaseModel
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (r *containerRegistryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("Container Registry"),
			"name":        common.SchemaResourceName("Container Registry"),
			"description": common.SchemaResourceDescription("Container Registry"),
			"tags":        common.SchemaResourceTags("Container Registry"),
			"icon_id":     common.SchemaResourceIconID("Container Registry"),
			"access_level": schema.StringAttribute{
				Required: true,
				Description: desc.Sprintf(
					"The level of access that allow to users. This must be one of [%s]",
					iaastypes.ContainerRegistryAccessLevelStrings,
				),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.ContainerRegistryAccessLevelStrings...),
				},
			},
			"subdomain_label": schema.StringAttribute{
				Required: true,
				Description: desc.Sprintf(
					"The label at the lowest of the FQDN used when be accessed from users. %s",
					desc.Length(1, 64),
				),
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"virtual_domain": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The alias for accessing the container registry",
			},
			"fqdn": schema.StringAttribute{
				Computed:    true,
				Description: "The FQDN for accessing the Container Registry. FQDN is built from `subdomain_label` + `.sakuracr.jp`",
			},
			"user": schema.SetNestedAttribute{
				Optional:    true,
				Description: "User accounts for accessing the container registry",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The user name used to authenticate remote access",
						},
						"password": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The password used to authenticate remote access",
						},
						"permission": schema.StringAttribute{
							Required: true,
							Description: desc.Sprintf(
								"The level of access that allow to the user. This must be one of [%s]",
								iaastypes.ContainerRegistryPermissionStrings,
							),
							Validators: []validator.String{
								stringvalidator.OneOf(iaastypes.ContainerRegistryPermissionStrings...),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *containerRegistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *containerRegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan containerRegistryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	builder := expandContainerRegistryBuilder(&plan, r.client, "")
	reg, err := builder.Build(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("creating SakuraCloud ContainerRegistry failed: %s", err))
		return
	}

	gotReg := getContainerRegistry(ctx, r.client, reg.ID, &resp.State, &resp.Diagnostics)
	if gotReg == nil {
		return
	}
	plan.updateState(ctx, r.client, gotReg, true, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *containerRegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state containerRegistryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reg := getContainerRegistry(ctx, r.client, common.SakuraCloudID(state.ID.ValueString()), &resp.State, &resp.Diagnostics)
	if reg == nil {
		return
	}
	state.updateState(ctx, r.client, reg, true, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *containerRegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan containerRegistryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	regOp := iaas.NewContainerRegistryOp(r.client)
	reg, err := regOp.Read(ctx, common.SakuraCloudID(plan.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("could not read SakuraCloud ContainerRegistry[%s]: %s", plan.ID.ValueString(), err))
		return
	}
	builder := expandContainerRegistryBuilder(&plan, r.client, reg.SettingsHash)
	builder.ID = reg.ID
	if _, err := builder.Build(ctx); err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating SakuraCloud ContainerRegistry[%s] failed: %s", plan.ID.ValueString(), err))
		return
	}

	gotReg := getContainerRegistry(ctx, r.client, reg.ID, &resp.State, &resp.Diagnostics)
	if gotReg == nil {
		return
	}
	plan.updateState(ctx, r.client, gotReg, true, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *containerRegistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state containerRegistryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	regOp := iaas.NewContainerRegistryOp(r.client)
	gotReg := getContainerRegistry(ctx, r.client, common.SakuraCloudID(state.ID.ValueString()), &resp.State, &resp.Diagnostics)
	if gotReg == nil {
		return
	}

	err := regOp.Delete(ctx, gotReg.ID)
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("deleting SakuraCloud ContainerRegistry[%s] failed: %s", state.ID.ValueString(), err))
	}

	resp.State.RemoveResource(ctx)
}

func expandContainerRegistryBuilder(d *containerRegistryResourceModel, c *common.APIClient, settingsHash string) *registryBuilder.Builder {
	return &registryBuilder.Builder{
		Name:           d.Name.ValueString(),
		Description:    d.Description.ValueString(),
		Tags:           common.TsetToStrings(d.Tags),
		IconID:         common.SakuraCloudID(d.IconID.ValueString()),
		AccessLevel:    iaastypes.EContainerRegistryAccessLevel(d.AccessLevel.ValueString()),
		VirtualDomain:  d.VirtualDomain.ValueString(),
		SubDomainLabel: d.SubDomainLabel.ValueString(),
		Users:          expandContainerRegistryUsers(d.User),
		SettingsHash:   settingsHash,
		Client:         iaas.NewContainerRegistryOp(c),
	}
}

func getContainerRegistry(ctx context.Context, client *common.APIClient, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.ContainerRegistry {
	regOp := iaas.NewContainerRegistryOp(client)
	reg, err := regOp.Read(ctx, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Get Container Registry Error", fmt.Sprintf("could not read SakuraCloud ContainerRegistry[%s]: %s", id, err))
		return nil
	}
	return reg
}
