// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_dedicated

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/apprun-dedicated-api-go/apis/v1"
	ver "github.com/sacloud/apprun-dedicated-api-go/apis/version"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	sacloudvalidator "github.com/sacloud/terraform-provider-sakura/internal/validator"
)

type verDataSource struct{ dataSourceClient }
type verDataSourceModel struct{ verModel }

var (
	_ datasource.DataSource              = &verDataSource{}
	_ datasource.DataSourceWithConfigure = &verDataSource{}
)

func NewVersionDataSource() datasource.DataSource { return &verDataSource{dataSourceNamed("version")} }

func (d *verDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, res *datasource.SchemaResponse) {
	res.Schema = schema.Schema{
		Description: "Information about an AppRun dedicated application version",
		Attributes: map[string]schema.Attribute{
			"application_id": func() (attr schema.StringAttribute) {
				attr = common.SchemaDataSourceId("application").(schema.StringAttribute)
				attr.Required = true
				attr.Computed = false
				attr.Optional = false
				attr.Validators = []validator.String{sacloudvalidator.UUIDValidator}
				return
			}(),
			// unlike other apprun dedicated resources, a version has no name.
			// it has version number in stead.
			"version": d.schemaVersion(),
			"cpu": schema.Int64Attribute{
				Computed:    true,
				Description: "The CPU limit in millicores (e.g., 1000 = 1 CPU)",
			},
			"memory": schema.Int64Attribute{
				Computed:    true,
				Description: "The memory limit in megabytes",
			},
			"scaling_mode": schema.StringAttribute{
				Computed:    true,
				Description: "The scaling mode (manual, autoscale)",
			},
			"fixed_scale": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "Number of nodes when scaling mode is `manual`",
			},
			"min_scale": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "Minimum number of nodes when scaling mode is `autoscale`",
			},
			"max_scale": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "Maximum number of nodes when scaling mode is `autoscale`",
			},
			"scale_in_threshold": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "When to scale in when scaling mode is `autoscale`",
			},
			"scale_out_threshold": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "When to scale out when scaling mode is `autoscale`",
			},
			"image": schema.StringAttribute{
				Computed:    true,
				Description: "The container image",
			},
			"registry_username": schema.StringAttribute{
				Computed:    true,
				Description: "Login username for container registry",
			},
			// note that password is intentionally not saved in the state
			"cmd": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "application command line i.e. the command and arguments",
			},
			"active_node_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of active nodes.  You might want to ignore_changes this field because it changes from time to time",
			},
			"created_at": d.schemaCreatedAt(),
			"exposed_ports": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Ports that the application exposes",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"target_port": schema.Int32Attribute{
							Computed:    true,
							Description: "The port that the application listens to",
						},
						"lb_port": schema.Int32Attribute{
							Computed:    true,
							Description: "The port that the load balancer listens to",
						},
						"use_lets_encrypt": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the load balancer uses Let's Encrypt (applicable only when `https`)",
						},
						"host": schema.SetAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Target `Host:` header value (only applicable when `http` or `https`)",
						},
						"health_check": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Health check configuration",
							Attributes: map[string]schema.Attribute{
								"path": schema.StringAttribute{
									Computed:    true,
									Description: "Health check endpoint",
								},
								"interval": schema.Int32Attribute{
									Computed:    true,
									Description: "Health check intervals in seconds",
								},
								"timeout": schema.Int32Attribute{
									Computed:    true,
									Description: "Time out in seconds until the health check fails",
								},
							},
						},
					},
				},
			},
			"env_vars": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Environment variables",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Computed:    true,
							Description: "Environment variable name",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value, or null if it is secret",
						},
						"secret": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the value is sensitive",
						},
					},
				},
			},
		},
	}
}

func (d *verDataSource) Read(ctx context.Context, req datasource.ReadRequest, res *datasource.ReadResponse) {
	var state verDataSourceModel
	res.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if res.Diagnostics.HasError() {
		return
	}

	aid, ver, err := state.ver()

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated application: %s", err))
		return
	}

	detail, err := d.api(aid).Read(ctx, ver)

	if err != nil {
		res.Diagnostics.AddError("Read: API Error", fmt.Sprintf("failed to read AppRun Dedicated application: %s", err))
		return
	}

	res.Diagnostics.Append(state.updateState(ctx, detail, aid)...)
	res.Diagnostics.Append(res.State.Set(ctx, &state)...)
}

func (d *verDataSource) api(applicationID appID) ver.VersionAPI {
	return ver.NewVersionOp(d.client, applicationID)
}

func (d *verDataSourceModel) ver() (aid appID, ver v1.ApplicationVersionNumber, err error) {
	aid, err = d.appId()
	ver = v1.ApplicationVersionNumber(d.Version.ValueInt32())
	return
}

func (*verDataSource) schemaVersion() schema.Int32Attribute {
	return schema.Int32Attribute{
		Required:    true,
		Optional:    false,
		Computed:    false,
		Validators:  []validator.Int32{int32validator.AtLeast(1)},
		Description: "The version number of the application",
	}
}
