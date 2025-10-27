// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package simple_mq

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/simplemq-api-go"
	"github.com/sacloud/simplemq-api-go/apis/v1/queue"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type simpleMQDataSource struct {
	client *queue.Client
}

var (
	_ datasource.DataSource              = &simpleMQDataSource{}
	_ datasource.DataSourceWithConfigure = &simpleMQDataSource{}
)

func NewSimpleMQDataSource() datasource.DataSource {
	return &simpleMQDataSource{}
}

func (d *simpleMQDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_simple_mq"
}

func (d *simpleMQDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.SimpleMqClient
}

type simpleMQDataSourceModel struct {
	simpleMqBaseModel
}

func (d *simpleMQDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          common.SchemaDataSourceId("SimpleMQ"),
			"name":        common.SchemaDataSourceName("SimpleMQ"),
			"description": common.SchemaDataSourceDescription("SimpleMQ"),
			"icon_id":     common.SchemaDataSourceIconID("SimpleMQ"),
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: desc.Sprintf("The tags of the SimpleMQ."),
			},
			"visibility_timeout_seconds": schema.Int64Attribute{
				Computed:    true,
				Description: "The duration in seconds that a message is invisible to others after being read from a queue. Default is 30 seconds.",
			},
			"expire_seconds": schema.Int64Attribute{
				Computed:    true,
				Description: "The duration in seconds that a message is stored in a queue. Default is 345600 seconds (4 days).",
			},
		},
	}
}

func (d *simpleMQDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data simpleMQDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	tags := common.TsetToStrings(data.Tags)
	if name == "" && len(tags) == 0 {
		resp.Diagnostics.AddError("Invalid Attribute", "Either name or tags must be specified.")
		return
	}

	queueOp := simplemq.NewQueueOp(d.client)
	qs, err := queueOp.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("could not find SakuraCloud SimpleMQ resource: %s", err))
		return
	}

	item, err := filterSimpleMQByNameOrTags(qs, name, tags)
	if err != nil {
		resp.Diagnostics.AddError("Not Found", err.Error())
		return
	}
	data.updateState(item)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterSimpleMQByNameOrTags(qs []queue.CommonServiceItem, name string, tags []string) (*queue.CommonServiceItem, error) {
	var match []*queue.CommonServiceItem
	for i, v := range qs {
		if name != "" && name != simplemq.GetQueueName(&v) {
			continue
		}
		tagsMatched := true
		for _, tagToFind := range tags {
			found := false
			for _, tag := range v.Tags {
				if tag == tagToFind {
					found = true
					break
				}
			}
			if !found {
				tagsMatched = false
				break
			}
		}
		if !tagsMatched {
			continue
		}
		match = append(match, &qs[i])
	}
	if len(match) == 0 {
		return nil, fmt.Errorf("no SimpleMQ resource found")
	}
	if len(match) > 1 {
		return nil, fmt.Errorf("multiple SimpleMQ resources found with the same condition. name=%q & tags=%v", name, tags)
	}
	return match[0], nil
}
