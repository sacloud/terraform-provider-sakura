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

package kms

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kms "github.com/sacloud/kms-api-go"
	v1 "github.com/sacloud/kms-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
)

type kmsDataSource struct {
	client *v1.Client
}

var (
	_ datasource.DataSource              = &kmsDataSource{}
	_ datasource.DataSourceWithConfigure = &kmsDataSource{}
)

func NewKmsDataSource() datasource.DataSource {
	return &kmsDataSource{}
}

func (d *kmsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kms"
}

func (d *kmsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.KmsClient
}

type kmsDataSourceModel struct {
	common.SakuraBaseModel
	KeyOrigin types.String `tfsdk:"key_origin"`
}

func (d *kmsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("KMS key"),
			"description": common.SchemaDataSourceDescription("KMS key"),
			"tags":        common.SchemaDataSourceTags("KMS key"),
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the KMS key.",
			},
			"key_origin": schema.StringAttribute{
				Computed:    true,
				Description: "The key origin of the KMS key.",
			},
		},
	}
}

func (d *kmsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data kmsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Name.IsNull() && data.ID.IsNull() {
		resp.Diagnostics.AddError("Missing Attribute", "Either 'id' or 'name' must be specified.")
		return
	}

	keyOp := kms.NewKeyOp(d.client)

	var key *v1.Key
	var err error
	if !data.Name.IsNull() {
		keys, err := keyOp.List(ctx)
		if err != nil {
			resp.Diagnostics.AddError("KMS List Error", fmt.Sprintf("could not find KMS resource: %s", err))
			return
		}
		key, err = FilterKMSByName(keys, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("KMS Filter Error", err.Error())
			return
		}
	} else {
		key, err = keyOp.Read(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("KMS Read Error", "No result found")
			return
		}
	}

	data.UpdateBaseState(key.ID, key.Name, key.Description.Value, key.Tags)
	data.KeyOrigin = types.StringValue(string(key.KeyOrigin))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FilterKMSByName(keys v1.Keys, name string) (*v1.Key, error) {
	match := slices.Collect(func(yield func(v1.Key) bool) {
		for _, v := range keys {
			if name != v.Name {
				continue
			}
			if !yield(v) {
				return
			}
		}
	})
	if len(match) == 0 {
		return nil, fmt.Errorf("no result")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple KMS resources found with the same condition. name=%q", name)
	}
	return &match[0], nil
}
