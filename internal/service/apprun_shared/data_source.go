// Copyright 2016-2025 The sacloud/terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_shared

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type apprunSharedDataSource struct {
	client *apprun.Client
}

var (
	_ datasource.DataSource              = &apprunSharedDataSource{}
	_ datasource.DataSourceWithConfigure = &apprunSharedDataSource{}
)

func NewApprunSharedDataSource() datasource.DataSource {
	return &apprunSharedDataSource{}
}

func (d *apprunSharedDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apprun_shared"
}

func (d *apprunSharedDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	d.client = apiclient.AppRunClient
}

type apprunSharedDataSourceModel struct {
	apprunSharedBaseModel
}

func (r *apprunSharedDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the AppRun Shared application",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of AppRun Shared application",
			},
			"timeout_seconds": schema.Int32Attribute{
				Computed:    true,
				Description: "The time limit between accessing the AppRun Shared application's public URL, starting the instance, and receiving a response",
			},
			"port": schema.Int32Attribute{
				Computed:    true,
				Description: "The port number where the AppRun Shared application listens for requests",
			},
			"min_scale": schema.Int32Attribute{
				Computed:    true,
				Description: "The minimum number of scales for the entire AppRun Shared application",
			},
			"max_scale": schema.Int32Attribute{
				Computed:    true,
				Description: "The maximum number of scales for the entire AppRun Shared application",
			},
			"components": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The AppRun Shared application component information",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The component name",
						},
						"max_cpu": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("The maximum number of CPUs for a component. The values in the list must be in [%s]", apprun.ApplicationMaxCPUs),
						},
						"max_memory": schema.StringAttribute{
							Computed:    true,
							Description: desc.Sprintf("The maximum memory of component. The values in the list must be in [%s]", apprun.ApplicationMaxMemories),
						},
						"deploy_source": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "The sources that make up the component",
							Attributes: map[string]schema.Attribute{
								"container_registry": schema.SingleNestedAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Container registry settings",
									Attributes: map[string]schema.Attribute{
										"image": schema.StringAttribute{
											Computed:    true,
											Description: "The container image name",
										},
										"server": schema.StringAttribute{
											Optional:    true,
											Computed:    true,
											Description: "The container registry server name",
										},
										"username": schema.StringAttribute{
											Optional:    true,
											Computed:    true,
											Description: "The container registry credentials",
										},
										"password": schema.StringAttribute{ // In data source, password is always empty string
											Computed:    true,
											Description: "The container registry credentials",
										},
									},
								},
							},
						},
						"env": schema.SetNestedAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The environment variables passed to components",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Computed:    true,
										Optional:    true,
										Description: "The environment variable name",
									},
									"value": schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Sensitive:   true,
										Description: "environment variable value",
									},
								},
							},
						},
						"probe": schema.SingleNestedAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The component probe settings",
							Attributes: map[string]schema.Attribute{
								"http_get": schema.SingleNestedAttribute{
									Computed:    true,
									Description: "HTTP probe settings",
									Attributes: map[string]schema.Attribute{
										"path": schema.StringAttribute{
											Computed:    true,
											Description: "The path to access HTTP server to check probes",
										},
										"port": schema.Int32Attribute{
											Computed:    true,
											Description: "The port number for accessing HTTP server and checking probes",
										},
										"headers": schema.SetNestedAttribute{
											Optional:    true,
											Computed:    true,
											Description: "HTTP headers for probe",
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Computed:    true,
														Description: "The header field name",
													},
													"value": schema.StringAttribute{
														Computed:    true,
														Description: "The header field value",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"packet_filter": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The packet filter for the AppRun Shared application",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed: true,
					},
					"settings": schema.ListNestedAttribute{
						Computed:    true,
						Description: "The list of packet filter rule",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"from_ip": schema.StringAttribute{
									Computed:    true,
									Description: "The source IP address of the rule",
								},
								"from_ip_prefix_length": schema.Int32Attribute{
									Computed:    true,
									Description: "The prefix length (CIDR notation) of the from_ip address, indicating the network size",
								},
							},
						},
					},
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The AppRun Shared application status",
			},
			"public_url": schema.StringAttribute{
				Computed:    true,
				Description: "The public URL of the AppRun Shared application",
			},
		},
	}
}

func (d *apprunSharedDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data apprunSharedDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appOp := apprun.NewApplicationOp(d.client)
	apps, err := appOp.List(ctx, &v1.ListApplicationsParams{})
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to list SakuraCloud AppRun Shared resource: %s", err))
		return
	}
	if apps == nil || len(apps.Data) == 0 {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	name := data.Name.ValueString()
	var app *v1.Application
	for _, d := range apps.Data {
		if d.Name == name {
			a, err := appOp.Read(ctx, d.Id)
			if err != nil {
				resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to read AppRun Shared resource: %s", err))
				return
			}
			app = a
			break
		}
	}
	if app == nil {
		common.FilterNoResultErr(&resp.Diagnostics)
		return
	}

	pfOp := apprun.NewPacketFilterOp(d.client)
	pf, err := pfOp.Read(ctx, app.Id)
	if err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to read AppRun Shared's PacketFilter resource: %s", err))
		return
	}

	data.updateState(app, pf)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
