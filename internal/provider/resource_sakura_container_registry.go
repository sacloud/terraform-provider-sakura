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

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	registryBuilder "github.com/sacloud/iaas-service-go/containerregistry/builder"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/desc"
)

// Terraform Plugin Framework Resource Model
type containerRegistryResourceModel struct {
	ID             types.String                  `tfsdk:"id"`
	Name           types.String                  `tfsdk:"name"`
	AccessLevel    types.String                  `tfsdk:"access_level"`
	VirtualDomain  types.String                  `tfsdk:"virtual_domain"`
	SubDomainLabel types.String                  `tfsdk:"subdomain_label"`
	FQDN           types.String                  `tfsdk:"fqdn"`
	IconID         types.String                  `tfsdk:"icon_id"`
	Description    types.String                  `tfsdk:"description"`
	Tags           types.Set                     `tfsdk:"tags"`
	User           []*containerRegistryUserModel `tfsdk:"user"`
}

type containerRegistryUserModel struct {
	Name       types.String `tfsdk:"name"`
	Password   types.String `tfsdk:"password"`
	Permission types.String `tfsdk:"permission"`
}

type containerRegistryResource struct {
	client *APIClient
}

func NewContainerRegistryResource() resource.Resource {
	return &containerRegistryResource{}
}

func (r *containerRegistryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_registry"
}

func (r *containerRegistryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiclient, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected ProviderData type", "Expected *APIClient.")
		return
	}

	r.client = apiclient
}

func (r *containerRegistryResource) Schema(_ context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaResourceId("Container Registry"),
			"name":        schemaResourceName("Container Registry"),
			"description": schemaResourceDescription("Container Registry"),
			"tags":        schemaResourceTags("Container Registry"),
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
			"icon_id": schemaResourceIconID("Container Registry"),
		},
		Blocks: map[string]schema.Block{
			"user": schema.SetNestedBlock{
				Description: "User accounts for accessing the container registry",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The user name used to authenticate remote access",
						},
						"password": schema.StringAttribute{
							Required: true,
							//Sensitive:   true,
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
		},
	}
}

func (r *containerRegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan containerRegistryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	reg := getContainerRegistry(ctx, r.client, sakuraCloudID(state.ID.ValueString()), &resp.State, &resp.Diagnostics)
	if reg == nil {
		return
	}
	state.updateState(ctx, r.client, reg, true, &resp.Diagnostics)
	/*
		regOp := iaas.NewContainerRegistryOp(r.client)
		reg, err := regOp.Read(ctx, sakuraCloudID(state.ID.ValueString()))
		if err != nil {
			if iaas.IsNotFoundError(err) {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Read Error", fmt.Sprintf("could not find SakuraCloud ContainerRegistry[%s]: %s", state.ID.ValueString(), err))
			return
		}


			state.ID = types.StringValue(reg.ID.String())
			state.Name = types.StringValue(reg.Name)
			state.AccessLevel = types.StringValue(string(reg.AccessLevel))
			state.VirtualDomain = types.StringValue(reg.VirtualDomain)
			state.SubDomainLabel = types.StringValue(reg.SubDomainLabel)
			state.FQDN = types.StringValue(reg.FQDN)
			state.IconID = types.StringValue(reg.IconID.String())
			state.Description = types.StringValue(reg.Description)
			state.Tags = stringsToTset(ctx, reg.Tags)
			state.User = flattenContainerRegistryUsers(state.User, getContainerRegistryUsers(ctx, r.client, reg), true)
	*/

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *containerRegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan containerRegistryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	regOp := iaas.NewContainerRegistryOp(r.client)
	reg, err := regOp.Read(ctx, sakuraCloudID(plan.ID.ValueString()))
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

	regOp := iaas.NewContainerRegistryOp(r.client)
	gotReg := getContainerRegistry(ctx, r.client, sakuraCloudID(state.ID.ValueString()), &resp.State, &resp.Diagnostics)
	if gotReg == nil {
		return
	}

	err := regOp.Delete(ctx, gotReg.ID)
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("deleting SakuraCloud ContainerRegistry[%s] failed: %s", state.ID.ValueString(), err))
	}

	resp.State.RemoveResource(ctx)
}

func expandContainerRegistryBuilder(d *containerRegistryResourceModel, c *APIClient, settingsHash string) *registryBuilder.Builder {
	return &registryBuilder.Builder{
		Name:           d.Name.ValueString(),
		Description:    d.Description.ValueString(),
		Tags:           tsetToStrings(d.Tags),
		IconID:         sakuraCloudID(d.IconID.ValueString()),
		AccessLevel:    iaastypes.EContainerRegistryAccessLevel(d.AccessLevel.ValueString()),
		VirtualDomain:  d.VirtualDomain.ValueString(),
		SubDomainLabel: d.SubDomainLabel.ValueString(),
		Users:          expandContainerRegistryUsers(d.User),
		SettingsHash:   settingsHash,
		Client:         iaas.NewContainerRegistryOp(c),
	}
}

func expandContainerRegistryUsers(users []*containerRegistryUserModel) []*registryBuilder.User {
	if len(users) == 0 {
		return nil
	}

	var results []*registryBuilder.User
	for _, u := range users {
		results = append(results, &registryBuilder.User{
			UserName:   u.Name.ValueString(),
			Password:   u.Password.ValueString(),
			Permission: iaastypes.EContainerRegistryPermission(u.Permission.ValueString()),
		})
	}
	return results
}

func getContainerRegistry(ctx context.Context, client *APIClient, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.ContainerRegistry {
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

func getContainerRegistryUsers(ctx context.Context, client *APIClient, user *iaas.ContainerRegistry) []*iaas.ContainerRegistryUser {
	regOp := iaas.NewContainerRegistryOp(client)
	users, err := regOp.ListUsers(ctx, user.ID)
	if err != nil {
		return nil
	}

	return users.Users
}

func flattenContainerRegistryUsers(conf []*containerRegistryUserModel, users []*iaas.ContainerRegistryUser, includePassword bool) []*containerRegistryUserModel {
	inputs := expandContainerRegistryUsers(conf)

	var results []*containerRegistryUserModel
	for _, user := range users {
		v := &containerRegistryUserModel{
			Name:       types.StringValue(user.UserName),
			Permission: types.StringValue(string(user.Permission)),
		}
		if includePassword {
			password := ""
			for _, i := range inputs {
				if i.UserName == user.UserName {
					password = i.Password
					break
				}
			}
			v.Password = types.StringValue(password)
		}
		results = append(results, v)
	}
	return results
}

func (model *containerRegistryResourceModel) updateState(ctx context.Context, c *APIClient, reg *iaas.ContainerRegistry, includePassword bool, diags *diag.Diagnostics) {
	users := getContainerRegistryUsers(ctx, c, reg)
	if users == nil {
		diags.AddError("Get Users Error", "could not get users for SakuraCloud ContainerRegistry")
		return
	}

	model.ID = types.StringValue(reg.ID.String())
	model.Name = types.StringValue(reg.Name)
	model.AccessLevel = types.StringValue(string(reg.AccessLevel))
	model.VirtualDomain = types.StringValue(reg.VirtualDomain)
	model.SubDomainLabel = types.StringValue(reg.SubDomainLabel)
	model.FQDN = types.StringValue(reg.FQDN)
	model.IconID = types.StringValue(reg.IconID.String())
	model.Description = types.StringValue(reg.Description)
	model.Tags = stringsToTset(ctx, reg.Tags)
	model.User = flattenContainerRegistryUsers(model.User, users, includePassword)
}
