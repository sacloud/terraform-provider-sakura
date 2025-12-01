// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package kms

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	kms "github.com/sacloud/kms-api-go"
	v1 "github.com/sacloud/kms-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
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
	kmsBaseModel
}

func (d *kmsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("KMS key"),
			"name":        common.SchemaDataSourceName("KMS key"),
			"description": common.SchemaDataSourceDescription("KMS key"),
			"tags":        common.SchemaDataSourceTags("KMS key"),
			"key_origin": schema.StringAttribute{
				Computed:    true,
				Description: "The key origin of the KMS key.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the KMS key.",
			},
			"latest_version": schema.Int64Attribute{
				Computed:    true,
				Description: "The latest material version of the KMS key.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The creation time of the KMS key.",
			},
			"modified_at": schema.StringAttribute{
				Computed:    true,
				Description: "The last modification time of the KMS key.",
			},
		},
		MarkdownDescription: "Get information about an existing KMS.",
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
			resp.Diagnostics.AddError("List Error", fmt.Sprintf("failed to find KMS resources: %s", err))
			return
		}
		key, err = FilterKMSByName(keys, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Filter Error", err.Error())
			return
		}
	} else {
		key, err = keyOp.Read(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to read KMS resource[%s]: %s", data.ID.ValueString(), err.Error()))
			return
		}
	}

	data.updateState(key)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func FilterKMSByName(keys []v1.Key, name string) (*v1.Key, error) {
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
