// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package ondemand_db

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
	"github.com/sacloud/iaas-service-go/enhanceddb/builder"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type onDemandDBResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &onDemandDBResource{}
	_ resource.ResourceWithConfigure   = &onDemandDBResource{}
	_ resource.ResourceWithImportState = &onDemandDBResource{}
)

func NewOnDemandDBResource() resource.Resource {
	return &onDemandDBResource{}
}

func (r *onDemandDBResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ondemand_db"
}

func (r *onDemandDBResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiClient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiClient == nil {
		return
	}
	r.client = apiClient
}

type onDemandDBResourceModel struct {
	onDemandDBBaseModel
	PasswordWO        types.String   `tfsdk:"password_wo"`
	PasswordWOVersion types.Int32    `tfsdk:"password_wo_version"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func (r *onDemandDBResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceName := "OnDemand Database"
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId(resourceName),
			"name":        common.SchemaResourceName(resourceName),
			"description": common.SchemaResourceDescription(resourceName),
			"tags":        common.SchemaResourceTags(resourceName),
			"icon_id":     common.SchemaResourceIconID(resourceName),
			"database_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of database",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"database_type": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The type of database. This must be one of [%s]", []string{"tidb"}),
				Validators: []validator.String{
					stringvalidator.OneOf("tidb"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Required:    true,
				Description: desc.Sprintf("The name of region that the database is in. This must be one of [%s]", iaastypes.EnhancedDBRegionStrings),
				Validators: []validator.String{
					stringvalidator.OneOf(iaastypes.EnhancedDBRegionStrings...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password_wo": schema.StringAttribute{
				Required:    true,
				WriteOnly:   true,
				Description: "The password of database",
				Validators: []validator.String{
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
			"allowed_networks": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of CIDR blocks allowed to connect",
			},
			"hostname": schema.StringAttribute{
				Computed:    true,
				Description: "The name of database host. This will be built from `database_name` + `tidb-is1.db.sakurausercontent.com`",
			},
			"max_connections": schema.Int64Attribute{
				Computed:    true,
				Description: "The value of max connections setting",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an OnDemand Database.",
	}
}

func (r *onDemandDBResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *onDemandDBResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config onDemandDBResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	edbBuilder := expandOnDemandDBBuilder(&plan, &config, r.client, "")
	created, err := edbBuilder.Build(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Create: API Error", fmt.Sprintf("failed to create OnDemand Database: %s", err))
		return
	}

	plan.updateState(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *onDemandDBResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state onDemandDBResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	edb := getOnDemandDB(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if edb == nil {
		return
	}

	state.updateState(edb)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *onDemandDBResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, config onDemandDBResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	reg := getOnDemandDB(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if reg == nil {
		return
	}

	edbBuilder := expandOnDemandDBBuilder(&plan, &config, r.client, reg.SettingsHash)
	if _, err := edbBuilder.Build(ctx); err != nil {
		resp.Diagnostics.AddError("Update: API Error", fmt.Sprintf("failed to update OnDemand Database[%s]: %s", plan.ID.ValueString(), err))
		return
	}

	edb := getOnDemandDB(ctx, r.client, reg.ID.String(), &resp.State, &resp.Diagnostics)
	if edb == nil {
		return
	}

	plan.updateState(edb)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *onDemandDBResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state onDemandDBResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	edb := getOnDemandDB(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if edb == nil {
		return
	}

	if err := iaas.NewEnhancedDBOp(r.client).Delete(ctx, edb.ID); err != nil {
		resp.Diagnostics.AddError("Delete: API Error", fmt.Sprintf("failed to delete OnDemand Database[%s]: %s", state.ID.ValueString(), err))
		return
	}
}

func getOnDemandDB(ctx context.Context, client *common.APIClient, id string, state *tfsdk.State, diags *diag.Diagnostics) *builder.EnhancedDB {
	edbOp := iaas.NewEnhancedDBOp(client)
	edb, err := builder.Read(ctx, edbOp, common.SakuraCloudID(id))
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Error", fmt.Sprintf("failed to read OnDemand Database[%s]: %s", id, err))
		return nil
	}
	return edb
}

func expandOnDemandDBBuilder(m, conf *onDemandDBResourceModel, client *common.APIClient, settingsHash string) *builder.Builder {
	return &builder.Builder{
		ID:              common.ExpandSakuraCloudID(m.ID),
		Name:            m.Name.ValueString(),
		Description:     m.Description.ValueString(),
		Tags:            common.TsetToStrings(m.Tags),
		IconID:          common.ExpandSakuraCloudID(m.IconID),
		DatabaseName:    m.DatabaseName.ValueString(),
		DatabaseType:    iaastypes.EnhancedDBType(m.DatabaseType.ValueString()),
		Region:          iaastypes.EnhancedDBRegion(m.Region.ValueString()),
		Password:        conf.PasswordWO.ValueString(),
		AllowedNetworks: common.TlistToStringsOrDefault(m.AllowedNetworks),
		SettingsHash:    settingsHash,
		Client:          iaas.NewEnhancedDBOp(client),
	}
}
