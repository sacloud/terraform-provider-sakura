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
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	api "github.com/sacloud/api-client-go"
	"github.com/sacloud/kms-api-go"
	v1 "github.com/sacloud/kms-api-go/apis/v1"
)

// Resource Model
type kmsResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	KeyOrigin   types.String `tfsdk:"key_origin"`
	PlainKey    types.String `tfsdk:"plain_key"`
	Description types.String `tfsdk:"description"`
	Tags        types.Set    `tfsdk:"tags"`
}

type kmsResource struct {
	client *v1.Client
}

func NewKMSResource() resource.Resource {
	return &kmsResource{}
}

func (r *kmsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiclient, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected ProviderData type", "Expected *APIClient.")
		return
	}

	r.client = apiclient.kmsClient
}

func (r *kmsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms"
}

func (r *kmsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schemaDataSourceId("KMS key"),
			"name":        schemaResourceName("KMS key"),
			"description": schemaResourceDescription("KMS key"),
			"tags":        schemaResourceTags("KMS key"),
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
		},
		// TODO: timeouts
	}
}

func (r *kmsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *kmsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data kmsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyReq, err := expandKMSCreateKey(&data)
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

	data.ID = types.StringValue(createdKey.ID)
	data.Name = types.StringValue(createdKey.Name)
	data.KeyOrigin = types.StringValue(string(createdKey.KeyOrigin))
	data.Description = types.StringValue(createdKey.Description.Value)
	data.Tags = stringsToTset(ctx, createdKey.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kmsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data kmsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyOp := kms.NewKeyOp(r.client)
	key, err := keyOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		if api.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("KMS Read Error", err.Error())
		return
	}

	data.ID = types.StringValue(key.ID)
	data.Name = types.StringValue(key.Name)
	data.KeyOrigin = types.StringValue(string(key.KeyOrigin))
	data.Description = types.StringValue(key.Description.Value)
	data.Tags = stringsToTset(ctx, key.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kmsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data kmsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyOp := kms.NewKeyOp(r.client)
	key, err := keyOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("KMS Update's Read Error", err.Error())
		return
	}

	_, err = keyOp.Update(ctx, key.ID, expandKMSUpdateKey(&data, key))
	if err != nil {
		resp.Diagnostics.AddError("KMS Update Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *kmsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data kmsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyOp := kms.NewKeyOp(r.client)
	key, err := keyOp.Read(ctx, data.ID.ValueString())
	if err != nil {
		if api.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("KMS Delete's Read Error", err.Error())
		return
	}

	if err := keyOp.Delete(ctx, key.ID); err != nil {
		resp.Diagnostics.AddError("KMS Delete Error", err.Error())
		return
	}
}

func expandKMSCreateKey(d *kmsResourceModel) (v1.CreateKey, error) {
	keyOrig := d.KeyOrigin.ValueString()
	var req v1.CreateKey
	if keyOrig == "generated" {
		req = v1.CreateKey{
			Name:      d.Name.ValueString(),
			KeyOrigin: v1.KeyOriginEnumGenerated,
		}
	} else {
		plainKey := d.PlainKey.ValueString()
		if plainKey == "" {
			return v1.CreateKey{}, errors.New("plain_key is required when key_origin is 'imported'")
		}
		req = v1.CreateKey{
			Name:      d.Name.ValueString(),
			KeyOrigin: v1.KeyOriginEnumImported,
			PlainKey:  v1.NewOptString(plainKey),
		}
	}

	if !d.Tags.IsNull() {
		req.Tags = tsetToStrings(d.Tags)
	}
	if !d.Description.IsNull() {
		req.Description = v1.NewOptString(d.Description.ValueString())
	}

	return req, nil
}

func expandKMSUpdateKey(d *kmsResourceModel, before *v1.Key) v1.Key {
	req := v1.Key{
		Name:      d.Name.ValueString(),
		KeyOrigin: before.KeyOrigin,
	}

	if !d.Tags.IsNull() {
		req.Tags = tsetToStrings(d.Tags)
	}
	if !d.Description.IsNull() {
		req.Description = v1.NewOptString(d.Description.ValueString())
	}

	return req
}
