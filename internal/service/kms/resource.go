// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package kms

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-plugin-framework/types"

	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/kms-api-go"
	v1 "github.com/sacloud/kms-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

type kmsResource struct {
	client *v1.Client
}

var (
	_ resource.Resource                = &kmsResource{}
	_ resource.ResourceWithConfigure   = &kmsResource{}
	_ resource.ResourceWithImportState = &kmsResource{}
)

func NewKMSResource() resource.Resource {
	return &kmsResource{}
}

func (r *kmsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms"
}

func (r *kmsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.KmsClient
}

type kmsResourceModel struct {
	common.SakuraBaseModel
	KeyOrigin types.String   `tfsdk:"key_origin"`
	PlainKey  types.String   `tfsdk:"plain_key"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

func (r *kmsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaResourceId("KMS key"),
			"name":        common.SchemaResourceName("KMS key"),
			"description": common.SchemaResourceDescription("KMS key"),
			"tags":        common.SchemaResourceTags("KMS key"),
			"key_origin": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("generated"),
				Description: "Key origin of the KMS key. 'generated' or 'imported'. Default is 'generated'.",
				Validators: []validator.String{
					stringvalidator.OneOf("generated", "imported"),
				},
			},
			"plain_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Plain key for imported KMS key. Required when `key_origin` is 'imported'.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *kmsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *kmsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan kmsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	keyReq, err := expandKMSCreateKey(&plan)
	if err != nil {
		resp.Diagnostics.AddError("KMS Create Key Expansion Error", err.Error())
		return
	}

	keyOp := kms.NewKeyOp(r.client)
	createdKey, err := keyOp.Create(ctx, keyReq)
	if err != nil {
		resp.Diagnostics.AddError("KMS Create Error", err.Error())
		return
	}

	plan.UpdateBaseState(createdKey.ID, createdKey.Name, createdKey.Description.Value, createdKey.Tags)
	plan.KeyOrigin = types.StringValue(string(createdKey.KeyOrigin))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kmsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data kmsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := getKMS(ctx, r.client, data.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	data.UpdateBaseState(key.ID, key.Name, key.Description.Value, key.Tags)
	data.KeyOrigin = types.StringValue(string(key.KeyOrigin))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kmsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan kmsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	keyOp := kms.NewKeyOp(r.client)
	key := getKMS(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	_, err := keyOp.Update(ctx, key.ID, expandKMSUpdateKey(&plan, key))
	if err != nil {
		resp.Diagnostics.AddError("KMS Update Error", err.Error())
		return
	}

	key = getKMS(ctx, r.client, plan.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	plan.UpdateBaseState(key.ID, key.Name, key.Description.Value, key.Tags)
	plan.KeyOrigin = types.StringValue(string(key.KeyOrigin))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kmsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state kmsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	keyOp := kms.NewKeyOp(r.client)
	key := getKMS(ctx, r.client, state.ID.ValueString(), &resp.State, &resp.Diagnostics)
	if key == nil {
		return
	}

	if err := keyOp.Delete(ctx, key.ID); err != nil {
		resp.Diagnostics.AddError("KMS Delete Error", err.Error())
		return
	}
}

func getKMS(ctx context.Context, client *v1.Client, id string, state *tfsdk.State, diags *diag.Diagnostics) *v1.Key {
	keyOp := kms.NewKeyOp(client)
	key, err := keyOp.Read(ctx, id)
	if err != nil {
		if api.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("Get KMS Key Error", fmt.Sprintf("could not read SakuraCloud KMS key[%s]: %s", id, err))
		return nil
	}

	return key
}

func expandKMSCreateKey(model *kmsResourceModel) (v1.CreateKey, error) {
	keyOrig := model.KeyOrigin.ValueString()
	var req v1.CreateKey
	if keyOrig == "generated" {
		req = v1.CreateKey{
			Name:      model.Name.ValueString(),
			KeyOrigin: v1.KeyOriginEnumGenerated,
		}
	} else {
		plainKey := model.PlainKey.ValueString()
		if plainKey == "" {
			return v1.CreateKey{}, errors.New("plain_key is required when key_origin is 'imported'")
		}
		req = v1.CreateKey{
			Name:      model.Name.ValueString(),
			KeyOrigin: v1.KeyOriginEnumImported,
			PlainKey:  v1.NewOptString(plainKey),
		}
	}

	if !model.Tags.IsNull() {
		req.Tags = common.TsetToStrings(model.Tags)
	}
	if !model.Description.IsNull() {
		req.Description = v1.NewOptString(model.Description.ValueString())
	}

	return req, nil
}

func expandKMSUpdateKey(model *kmsResourceModel, before *v1.Key) v1.Key {
	req := v1.Key{
		Name:      model.Name.ValueString(),
		KeyOrigin: before.KeyOrigin,
	}

	if !model.Tags.IsNull() {
		req.Tags = common.TsetToStrings(model.Tags)
	}
	if !model.Description.IsNull() {
		req.Description = v1.NewOptString(model.Description.ValueString())
	}

	return req
}
