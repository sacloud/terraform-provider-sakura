// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package icon

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mitchellh/go-homedir"
	"github.com/sacloud/iaas-api-go"
	iaastypes "github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type iconResource struct {
	client *common.APIClient
}

var (
	_ resource.Resource                = &iconResource{}
	_ resource.ResourceWithConfigure   = &iconResource{}
	_ resource.ResourceWithImportState = &iconResource{}
)

func NewIconResource() resource.Resource {
	return &iconResource{}
}

func (r *iconResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_icon"
}

func (r *iconResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient
}

type iconResourceModel struct {
	iconBaseModel
	Source        types.String   `tfsdk:"source"`
	Base64Content types.String   `tfsdk:"base64content"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

func (r *iconResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":   common.SchemaResourceId("Icon"),
			"name": common.SchemaResourceName("Icon"),
			"tags": common.SchemaResourceTags("Icon"),
			"source": schema.StringAttribute{
				Optional:    true,
				Description: "The file path to upload as the Icon.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("base64content")),
				},
			},
			"base64content": schema.StringAttribute{
				Optional:    true,
				Description: "The base64 encoded content to upload as the Icon.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("source")),
				},
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL for getting the icon's raw data.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
		MarkdownDescription: "Manages an Icon.",
	}
}

func (r *iconResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *iconResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan iconResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	iconOp := iaas.NewIconOp(r.client)
	createReq, err := expandIconCreateRequest(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Icon Create Request Error", err.Error())
		return
	}
	icon, err := iconOp.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Icon Create API Error", err.Error())
		return
	}

	plan.updateState(icon)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *iconResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iconResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	icon := getIcon(ctx, r.client, common.SakuraCloudID(state.ID.ValueString()), &resp.State, &resp.Diagnostics)
	if icon == nil {
		return
	}

	state.updateState(icon)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *iconResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan iconResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	iconOp := iaas.NewIconOp(r.client)
	_, err := iconOp.Update(ctx, common.ExpandSakuraCloudID(plan.ID), expandIconUpdateRequest(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Icon Update API Error", err.Error())
		return
	}

	icon := getIcon(ctx, r.client, common.ExpandSakuraCloudID(plan.ID), &resp.State, &resp.Diagnostics)
	if icon == nil {
		return
	}

	plan.updateState(icon)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *iconResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iconResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	iconOp := iaas.NewIconOp(r.client)
	icon := getIcon(ctx, r.client, common.SakuraCloudID(state.ID.ValueString()), &resp.State, &resp.Diagnostics)
	if icon == nil {
		return
	}
	if err := iconOp.Delete(ctx, icon.ID); err != nil {
		resp.Diagnostics.AddError("Icon Delete API Error", err.Error())
		return
	}
}

func getIcon(ctx context.Context, client *common.APIClient, id iaastypes.ID, state *tfsdk.State, diags *diag.Diagnostics) *iaas.Icon {
	iconOp := iaas.NewIconOp(client)
	icon, err := iconOp.Read(ctx, id)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Icon Read API Error", err.Error())
		return nil
	}
	return icon
}

func expandIconBody(d *iconResourceModel) (string, error) {
	var body string
	if !d.Source.IsNull() { //nolint:gocritic
		source := d.Source.ValueString()
		path, err := homedir.Expand(source)
		if err != nil {
			return "", fmt.Errorf("expanding homedir in source (%s) is failed: %s", source, err)
		}
		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			return "", fmt.Errorf("opening SakuraCloud Icon source(%s) is failed: %s", source, err)
		}
		data, err := io.ReadAll(file)
		if err != nil {
			return "", fmt.Errorf("reading SakuraCloud Icon source file is failed: %s", err)
		}
		body = base64.StdEncoding.EncodeToString(data)
	} else if !d.Base64Content.IsNull() {
		body = d.Base64Content.ValueString()
	} else {
		return "", fmt.Errorf(`"source" or "base64content" field is required`)
	}
	return body, nil
}

func expandIconCreateRequest(d *iconResourceModel) (*iaas.IconCreateRequest, error) {
	body, err := expandIconBody(d)
	if err != nil {
		return nil, fmt.Errorf("creating SakuraCloud Icon is failed: %s", err)
	}
	return &iaas.IconCreateRequest{
		Name:  d.Name.ValueString(),
		Tags:  common.TsetToStrings(d.Tags),
		Image: body,
	}, nil
}

func expandIconUpdateRequest(d *iconResourceModel) *iaas.IconUpdateRequest {
	return &iaas.IconUpdateRequest{
		Name: d.Name.ValueString(),
		Tags: common.TsetToStrings(d.Tags),
	}
}
