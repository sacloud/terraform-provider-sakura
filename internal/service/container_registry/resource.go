// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package container_registry

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
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
	containerRegistryBaseModel
	User     []*containerRegistryUserModel `tfsdk:"user"`
	Timeouts timeouts.Value                `tfsdk:"timeouts"`
}

type containerRegistryUserModel struct {
	Name              types.String `tfsdk:"name"`
	Password          types.String `tfsdk:"password"`
	PasswordWO        types.String `tfsdk:"password_wo"`
	PasswordWOVersion types.Int32  `tfsdk:"password_wo_version"`
	Permission        types.String `tfsdk:"permission"`
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
				Optional: true,
				Computed: true,
				DeprecationMessage: "The \"access_level\" attribute is deprecated and will be removed in a future version. " +
					"Container Registry no longer supports public access settings. " +
					"See: https://cloud.sakura.ad.jp/news/2026/05/27/container-registry-public-access-setting-discontinued/",
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
				Description: "The alias for accessing the Container Registry",
			},
			"fqdn": schema.StringAttribute{
				Computed:    true,
				Description: "The FQDN for accessing the Container Registry. FQDN is built from `subdomain_label` + `.sakuracr.jp`",
			},
			// Setではwrite-onlyが使えないため、Listにする。CR API経由でのユーザ取得は順序が不定なため、レスポンスはチェックせず、configの値をそのまま使う。
			"user": schema.ListNestedAttribute{
				Optional:    true,
				Description: "User accounts for accessing the Container Registry",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The user name used to authenticate remote access",
						},
						"password": schema.StringAttribute{
							Optional:    true,
							Sensitive:   true,
							Description: "The password used to authenticate remote access",
							Validators: []validator.String{
								stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("user").AtAnyListIndex().AtName("password_wo")),
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password_wo")),
							},
						},
						"password_wo": schema.StringAttribute{
							Optional:    true,
							WriteOnly:   true,
							Description: "The password used to authenticate remote access",
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password")),
								stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo_version")),
							},
						},
						"password_wo_version": schema.Int32Attribute{
							Optional:    true,
							Description: "The version of the password_wo field. This value must be greater than 0 when set. Increment this when changing password.",
							Validators: []validator.Int32{
								int32validator.AtLeast(1),
								int32validator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo")),
							},
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
		MarkdownDescription: "Manages a Container Registry.",
	}
}

func (r *containerRegistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *containerRegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config containerRegistryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	builder := expandContainerRegistryBuilder(&plan, &config, r.client, "")
	reg, err := builder.Build(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create SakuraCloud Container Registry: %s", err))
		return
	}

	gotReg := getContainerRegistry(ctx, r.client, reg.ID, &resp.State, &resp.Diagnostics)
	if gotReg == nil {
		return
	}

	plan.updateState(gotReg)
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
	state.updateState(reg)
	if len(state.User) == 0 {
		users := getContainerRegistryUsers(ctx, r.client, reg)
		if users == nil {
			resp.Diagnostics.AddError("Read: API Error", "failed to get users for Container Registry")
			return
		}
		if len(users) > 0 {
			state.User = flattenContainerRegistryUsers(users)
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *containerRegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config containerRegistryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	regOp := iaas.NewContainerRegistryOp(r.client)
	reg, err := regOp.Read(ctx, common.SakuraCloudID(plan.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to read SakuraCloud Container Registry[%s]: %s", plan.ID.ValueString(), err))
		return
	}
	builder := expandContainerRegistryBuilder(&plan, &config, r.client, reg.SettingsHash)
	builder.ID = reg.ID
	if _, err := builder.Build(ctx); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update SakuraCloud Container Registry[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	gotReg := getContainerRegistry(ctx, r.client, reg.ID, &resp.State, &resp.Diagnostics)
	if gotReg == nil {
		return
	}
	plan.updateState(gotReg)
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

	gotReg := getContainerRegistry(ctx, r.client, common.SakuraCloudID(state.ID.ValueString()), &resp.State, &resp.Diagnostics)
	if gotReg == nil {
		return
	}

	regOp := iaas.NewContainerRegistryOp(r.client)
	err := regOp.Delete(ctx, gotReg.ID)
	if err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete SakuraCloud Container Registry[%s]: %s", state.ID.ValueString(), err))
	}
}

func expandContainerRegistryBuilder(d, config *containerRegistryResourceModel, c *common.APIClient, settingsHash string) *registryBuilder.Builder {
	accessLevel := iaastypes.ContainerRegistryAccessLevels.None
	if !d.AccessLevel.IsNull() && !d.AccessLevel.IsUnknown() {
		accessLevel = iaastypes.EContainerRegistryAccessLevel(d.AccessLevel.ValueString())
	}

	return &registryBuilder.Builder{
		Name:           d.Name.ValueString(),
		Description:    d.Description.ValueString(),
		Tags:           common.TsetToStrings(d.Tags),
		IconID:         common.SakuraCloudID(d.IconID.ValueString()),
		AccessLevel:    accessLevel,
		VirtualDomain:  d.VirtualDomain.ValueString(),
		SubDomainLabel: d.SubDomainLabel.ValueString(),
		Users:          expandContainerRegistryUsers(d.User, config.User),
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
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Container Registry[%s]: %s", id, err))
		return nil
	}

	return reg
}

func flattenContainerRegistryUsers(users []*iaas.ContainerRegistryUser) []*containerRegistryUserModel {
	var results []*containerRegistryUserModel
	for _, user := range users {
		v := &containerRegistryUserModel{
			Name:              types.StringValue(user.UserName),
			Password:          types.StringNull(),
			PasswordWO:        types.StringNull(),
			PasswordWOVersion: types.Int32Null(),
			Permission:        types.StringValue(string(user.Permission)),
		}
		results = append(results, v)
	}
	return results
}
