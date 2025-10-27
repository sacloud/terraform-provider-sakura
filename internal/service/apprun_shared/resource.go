// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package apprun_shared

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/sacloud/apprun-api-go"
	v1 "github.com/sacloud/apprun-api-go/apis/v1"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
)

type apprunSharedResource struct {
	client *apprun.Client
}

var (
	_ resource.Resource                = &apprunSharedResource{}
	_ resource.ResourceWithConfigure   = &apprunSharedResource{}
	_ resource.ResourceWithImportState = &apprunSharedResource{}
)

func NewApprunSharedResource() resource.Resource {
	return &apprunSharedResource{}
}

func (r *apprunSharedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apprun_shared"
}

func (r *apprunSharedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return
	}
	r.client = apiclient.AppRunClient
}

type apprunSharedResourceModel struct {
	apprunSharedBaseModel
	Traffics types.List     `tfsdk:"traffics"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type apprunSharedTrafficsModel struct {
	VersionIndex types.Int64 `tfsdk:"version_index"`
	Percent      types.Int32 `tfsdk:"percent"`
}

func (m apprunSharedTrafficsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"version_index": types.Int64Type,
		"percent":       types.Int32Type,
	}
}

func (r *apprunSharedResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the AppRun Shared",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": common.SchemaResourceName("AppRun Shared"),
			"timeout_seconds": schema.Int32Attribute{
				Required:    true,
				Description: "The time limit between accessing the AppRun Shared application's public URL, starting the instance, and receiving a response",
			},
			"port": schema.Int32Attribute{
				Required:    true,
				Description: "The port number where the AppRun Shared application listens for requests",
			},
			"min_scale": schema.Int32Attribute{
				Required:    true,
				Description: "The minimum number of scales for the entire AppRun Shared application",
			},
			"max_scale": schema.Int32Attribute{
				Required:    true,
				Description: "The maximum number of scales for the entire AppRun Shared application",
			},
			"components": schema.ListNestedAttribute{
				Required:    true,
				Description: "The AppRun Shared application component information",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The component name",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"max_cpu": schema.StringAttribute{
							Required:    true,
							Description: desc.Sprintf("The maximum number of CPUs for a component. The values in the list must be in [%s]", apprun.ApplicationMaxCPUs),
							Validators: []validator.String{
								stringvalidator.OneOf(apprun.ApplicationMaxCPUs...),
							},
						},
						"max_memory": schema.StringAttribute{
							Required:    true,
							Description: desc.Sprintf("The maximum memory of component. The values in the list must be in [%s]", apprun.ApplicationMaxMemories),
							Validators: []validator.String{
								stringvalidator.OneOf(apprun.ApplicationMaxMemories...),
							},
						},
						"deploy_source": schema.SingleNestedAttribute{
							Required:    true,
							Description: "The sources that make up the component",
							Attributes: map[string]schema.Attribute{
								"container_registry": schema.SingleNestedAttribute{
									// API側は他のソースを将来的にサポートする可能性があるためOptionalになっているが、現状はcontainer_registryのみサポートしている。
									// Terraformでは設定エラーを早期に発見するためにRequiredにし、将来的に他のソースがサポートされた時に改めてOptionalにする
									// https://manual.sakura.ad.jp/api/cloud/apprun/#tag/アプリケーション/operation/postApplication
									Required:    true,
									Description: "Container registry settings",
									Attributes: map[string]schema.Attribute{
										"image": schema.StringAttribute{
											Required:    true,
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
										"password": schema.StringAttribute{
											Optional:    true,
											Computed:    true,
											Sensitive:   true,
											Description: "The container registry credentials",
										},
									},
								},
							},
						},
						"env": schema.SetNestedAttribute{
							Optional:    true,
							Description: "The environment variables passed to components",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Required:    true,
										Description: "The environment variable name",
									},
									"value": schema.StringAttribute{
										Required:    true,
										Sensitive:   true,
										Description: "environment variable value",
									},
								},
							},
						},
						"probe": schema.SingleNestedAttribute{
							Optional:    true,
							Description: "The component probe settings",
							Attributes: map[string]schema.Attribute{
								"http_get": schema.SingleNestedAttribute{
									Required:    true,
									Description: "HTTP probe settings",
									Attributes: map[string]schema.Attribute{
										"path": schema.StringAttribute{
											Required:    true,
											Description: "The path to access HTTP server to check probes",
										},
										"port": schema.Int32Attribute{
											Required:    true,
											Description: "The port number for accessing HTTP server and checking probes",
										},
										"headers": schema.SetNestedAttribute{
											Optional:    true,
											Description: "HTTP headers for probe",
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"name": schema.StringAttribute{
														Required:    true,
														Description: "The header field name",
													},
													"value": schema.StringAttribute{
														Required:    true,
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
			"traffics": schema.ListNestedAttribute{
				Optional:    true,
				Description: "The AppRun Shared application traffic",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"version_index": schema.Int64Attribute{
							Required:    true,
							Description: "The AppRun Shared application version index",
						},
						"percent": schema.Int32Attribute{
							Required:    true,
							Description: "The percentage of traffic dispersion",
						},
					},
				},
			},
			"packet_filter": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The packet filter for the AppRun Shared application",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Required: true,
					},
					"settings": schema.ListNestedAttribute{
						Required:    true,
						Description: "The list of packet filter rule",
						Validators: []validator.List{
							listvalidator.SizeAtMost(5),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"from_ip": schema.StringAttribute{
									Required:    true,
									Description: "The source IP address of the rule",
								},
								"from_ip_prefix_length": schema.Int32Attribute{
									Required:    true,
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
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *apprunSharedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *apprunSharedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apprunSharedResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := createUserIfNotExist(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to create AppRun Shared user: %s", err))
		return
	}

	appOp := apprun.NewApplicationOp(r.client)
	params := v1.PostApplicationBody{
		Name:           plan.Name.ValueString(),
		TimeoutSeconds: int(plan.TimeoutSeconds.ValueInt32()),
		Port:           int(plan.Port.ValueInt32()),
		MinScale:       int(plan.MinScale.ValueInt32()),
		MaxScale:       int(plan.MaxScale.ValueInt32()),
		Components:     expandApprunApplicationComponents(&plan),
	}
	result, err := appOp.Create(ctx, &params)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to create AppRun Shared application: %s", err))
		return
	}

	pfOp := apprun.NewPacketFilterOp(r.client)
	if _, err := pfOp.Update(ctx, result.Id, expandApprunPacketFilter(&plan)); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to update AppRun Shared's packet filter: %s", err))
		return
	}

	// 内部的にVersions/Traffics APIを利用してトラフィック分散の状態も変更する
	versions, err := getVersions(ctx, r.client, result.Id)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to get AppRun Shared's versions: %s", err))
		return
	}

	trafficOp := apprun.NewTrafficOp(r.client)
	traffics, err := expandApprunApplicationTraffics(&plan, versions)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to expand AppRun Shared's traffics: %s", err))
		return
	}

	_, err = trafficOp.Update(ctx, result.Id, traffics)
	if err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to update AppRun Shared's traffics: %s", err))
		return
	}

	if _, err := updateModel(ctx, &plan, r.client, result.Id); err != nil {
		resp.Diagnostics.AddError("Create Error", fmt.Sprintf("failed to update AppRun Shared's state: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apprunSharedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apprunSharedResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := createUserIfNotExist(ctx, r.client); err != nil {
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to create AppRun Shared user: %s", err))
		return
	}

	if removeResource, err := updateModel(ctx, &state, r.client, state.ID.ValueString()); err != nil {
		if removeResource {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Read Error", fmt.Sprintf("failed to update AppRun Shared's state: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apprunSharedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apprunSharedResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appOp := apprun.NewApplicationOp(r.client)
	application, err := appOp.Read(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("read error", fmt.Errorf("failed to read AppRun Shared application: %s", err).Error())
		return
	}

	timeoutSeconds := int(plan.TimeoutSeconds.ValueInt32())
	port := int(plan.Port.ValueInt32())
	minScale := int(plan.MinScale.ValueInt32())
	maxScale := int(plan.MaxScale.ValueInt32())
	params := v1.PatchApplicationBody{
		TimeoutSeconds: &timeoutSeconds,
		Port:           &port,
		MinScale:       &minScale,
		MaxScale:       &maxScale,
		Components:     expandApprunApplicationComponentsForUpdate(&plan),
	}
	result, err := appOp.Update(ctx, application.Id, &params)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update AppRun Shared application: %s", err))
		return
	}

	pfOp := apprun.NewPacketFilterOp(r.client)
	if _, err := pfOp.Update(ctx, result.Id, expandApprunPacketFilter(&plan)); err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update AppRun Shared's packet filter: %s", err))
		return
	}

	versions, err := getVersions(ctx, r.client, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to get AppRun Shared's versions: %s", err))
		return
	}

	trafficOp := apprun.NewTrafficOp(r.client)
	traffics, err := expandApprunApplicationTraffics(&plan, versions)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to expand AppRun Shared's traffics: %s", err))
		return
	}

	_, err = trafficOp.Update(ctx, plan.ID.ValueString(), traffics)
	if err != nil {
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update AppRun Shared's traffics: %s", err))
		return
	}

	if removeResource, err := updateModel(ctx, &plan, r.client, result.Id); err != nil {
		if removeResource {
			resp.State.RemoveResource(ctx)
		}
		resp.Diagnostics.AddError("Update Error", fmt.Sprintf("failed to update AppRun Shared's state: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apprunSharedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apprunSharedResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appOp := apprun.NewApplicationOp(r.client)
	application, err := appOp.Read(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Errorf("failed to read AppRun Shared application: %s", err).Error())
		return
	}

	if err := appOp.Delete(ctx, application.Id); err != nil {
		resp.Diagnostics.AddError("Delete Error", fmt.Sprintf("failed to delete AppRun Shared application: %s", err))
		return
	}
}

func getVersions(ctx context.Context, client *apprun.Client, applicationId string) ([]v1.Version, error) {
	var versions []v1.Version

	versionOp := apprun.NewVersionOp(client)

	pageNum := 1
	pageSize := 100
	for {
		vs, err := versionOp.List(ctx, applicationId, &v1.ListApplicationVersionsParams{
			PageNum:  &pageNum,
			PageSize: &pageSize,
		})
		if err != nil {
			return nil, err
		}
		if len(vs.Data) == 0 {
			break
		}

		versions = append(versions, vs.Data...)
		pageNum++
	}

	return versions, nil
}

// NOTE: AppRunは初回利用時に一度のみユーザーの作成を必要とする。
// SakuraCloud Providerでは明示的にユーザーの作成を行わず、CURD操作の開始時に暗黙的にユーザーの存在確認と作成を行う。
// ref. https://manual.sakura.ad.jp/sakura-apprun-api/spec.html#tag/%E3%83%A6%E3%83%BC%E3%82%B8%E3%83%A3
func createUserIfNotExist(ctx context.Context, client *apprun.Client) error {
	userOp := apprun.NewUserOp(client)
	res, err := userOp.Read(ctx)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusNotFound {
		_, err := userOp.Create(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateModel(ctx context.Context, model *apprunSharedResourceModel, client *apprun.Client, id string) (bool, error) {
	appOp := apprun.NewApplicationOp(client)
	application, err := appOp.Read(ctx, id)
	if err != nil {
		if e, ok := err.(*v1.ModelDefaultError); ok && e.Detail.Code == 404 {
			return true, nil
		}
		return false, fmt.Errorf("failed to read AppRun Shared application: %s", err)
	}

	versions, err := getVersions(ctx, client, id)
	if err != nil {
		return false, fmt.Errorf("failed to get AppRun Shared's versions: %s", err)
	}

	trafficOp := apprun.NewTrafficOp(client)
	traffics, err := trafficOp.List(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to read AppRun Shared's traffics: %s", err)
	}

	pfOp := apprun.NewPacketFilterOp(client)
	pf, err := pfOp.Read(ctx, application.Id)
	if err != nil {
		return false, fmt.Errorf("failed to read AppRun Shared's packet filter: %s", err)
	}

	model.updateState(application, pf)
	model.Traffics = flattenApprunApplicationTraffics(traffics.Data, versions)
	return false, nil
}

func expandApprunApplicationComponentsForUpdate(model *apprunSharedResourceModel) *[]v1.PatchApplicationBodyComponent {
	var components []v1.PatchApplicationBodyComponent
	for _, component := range model.Components {
		// Create ContainerRegistry
		cr := component.DeploySource.ContainerRegistry
		containerRegistry := &v1.PatchApplicationBodyComponentDeploySourceContainerRegistry{
			Image: cr.Image.ValueString(),
		}
		if !cr.Server.IsNull() && !cr.Server.IsUnknown() {
			v := cr.Server.ValueString()
			containerRegistry.Server = &v
		}
		if !cr.Username.IsNull() && !cr.Username.IsUnknown() {
			v := cr.Username.ValueString()
			containerRegistry.Username = &v
		}
		if !cr.Password.IsNull() && !cr.Password.IsUnknown() {
			v := cr.Password.ValueString()
			containerRegistry.Password = &v
		}

		// Create Env
		envModel := make([]apprunSharedComponentEnvModel, 0, len(component.Env.Elements()))
		_ = component.Env.ElementsAs(context.Background(), &envModel, false)
		var env []v1.PatchApplicationBodyComponentEnv
		for _, e := range envModel {
			key := e.Key.ValueString()
			value := e.Value.ValueString()

			env = append(env,
				v1.PatchApplicationBodyComponentEnv{
					Key:   &key,
					Value: &value,
				})
		}

		// CreateProbe
		var probe v1.PatchApplicationBodyComponentProbe
		if !component.Probe.IsNull() && !component.Probe.IsUnknown() {
			var d apprunSharedProbeModel
			_ = component.Probe.As(context.Background(), &d, basetypes.ObjectAsOptions{})
			probe.HttpGet = &v1.PatchApplicationBodyComponentProbeHttpGet{
				Path: d.HttpGet.Path.ValueString(),
				Port: int(d.HttpGet.Port.ValueInt32()),
			}

			if !d.HttpGet.Headers.IsNull() && !d.HttpGet.Headers.IsUnknown() {
				headersModel := make([]apprunSharedProbeHttpGetHeaderModel, 0, len(d.HttpGet.Headers.Elements()))
				_ = d.HttpGet.Headers.ElementsAs(context.Background(), &headersModel, false)
				var headers []v1.PatchApplicationBodyComponentProbeHttpGetHeader
				for _, h := range headersModel {
					name := h.Name.ValueString()
					value := h.Value.ValueString()
					headers = append(headers,
						v1.PatchApplicationBodyComponentProbeHttpGetHeader{
							Name:  &name,
							Value: &value,
						})
				}

				probe.HttpGet.Headers = &headers
			}
		}

		components = append(components, v1.PatchApplicationBodyComponent{
			Name:      component.Name.ValueString(),
			MaxCpu:    v1.PatchApplicationBodyComponentMaxCpu(component.MaxCpu.ValueString()),
			MaxMemory: v1.PatchApplicationBodyComponentMaxMemory(component.MaxMemory.ValueString()),
			DeploySource: v1.PatchApplicationBodyComponentDeploySource{
				ContainerRegistry: containerRegistry,
			},
			Env:   &env,
			Probe: &probe,
		})
	}

	return &components
}

func expandApprunApplicationComponents(model *apprunSharedResourceModel) []v1.PostApplicationBodyComponent {
	var components []v1.PostApplicationBodyComponent
	for _, component := range model.Components {
		// Create ContainerRegistry
		cr := component.DeploySource.ContainerRegistry
		containerRegistry := &v1.PostApplicationBodyComponentDeploySourceContainerRegistry{
			Image: cr.Image.ValueString(),
		}
		if !cr.Server.IsNull() && !cr.Server.IsUnknown() {
			v := cr.Server.ValueString()
			containerRegistry.Server = &v
		}
		if !cr.Username.IsNull() && !cr.Username.IsUnknown() {
			v := cr.Username.ValueString()
			containerRegistry.Username = &v
		}
		if !cr.Password.IsNull() && !cr.Password.IsUnknown() {
			v := cr.Password.ValueString()
			containerRegistry.Password = &v
		}

		// Create Env
		envModel := make([]apprunSharedComponentEnvModel, 0, len(component.Env.Elements()))
		_ = component.Env.ElementsAs(context.Background(), &envModel, false)
		var env []v1.PostApplicationBodyComponentEnv
		for _, e := range envModel {
			key := e.Key.ValueString()
			value := e.Value.ValueString()

			env = append(env,
				v1.PostApplicationBodyComponentEnv{
					Key:   &key,
					Value: &value,
				})
		}

		// Create Probe
		var probe v1.PostApplicationBodyComponentProbe
		if !component.Probe.IsNull() && !component.Probe.IsUnknown() {
			var d apprunSharedProbeModel
			_ = component.Probe.As(context.Background(), &d, basetypes.ObjectAsOptions{})
			probe.HttpGet = &v1.PostApplicationBodyComponentProbeHttpGet{
				Path: d.HttpGet.Path.ValueString(),
				Port: int(d.HttpGet.Port.ValueInt32()),
			}

			if !d.HttpGet.Headers.IsNull() && !d.HttpGet.Headers.IsUnknown() {
				headersModel := make([]apprunSharedProbeHttpGetHeaderModel, 0, len(d.HttpGet.Headers.Elements()))
				_ = d.HttpGet.Headers.ElementsAs(context.Background(), &headersModel, false)
				var headers []v1.PostApplicationBodyComponentProbeHttpGetHeader
				for _, h := range headersModel {
					name := h.Name.ValueString()
					value := h.Value.ValueString()
					headers = append(headers,
						v1.PostApplicationBodyComponentProbeHttpGetHeader{
							Name:  &name,
							Value: &value,
						})
				}
				probe.HttpGet.Headers = &headers
			}
		}

		components = append(components, v1.PostApplicationBodyComponent{
			Name:      component.Name.ValueString(),
			MaxCpu:    v1.PostApplicationBodyComponentMaxCpu(component.MaxCpu.ValueString()),
			MaxMemory: v1.PostApplicationBodyComponentMaxMemory(component.MaxMemory.ValueString()),
			DeploySource: v1.PostApplicationBodyComponentDeploySource{
				ContainerRegistry: containerRegistry,
			},
			Env:   &env,
			Probe: &probe,
		})
	}

	return components
}

func expandApprunApplicationTraffics(model *apprunSharedResourceModel, versions []v1.Version) (*[]v1.Traffic, error) {
	// resourceにtraffics listが存在しない場合
	if model.Traffics.IsNull() || model.Traffics.IsUnknown() || len(model.Traffics.Elements()) == 0 {
		defaultIsLatestVersion := true
		defaultPercent := 100

		t := &v1.Traffic{}
		if err := t.FromTrafficWithLatestVersion(v1.TrafficWithLatestVersion{
			IsLatestVersion: defaultIsLatestVersion,
			Percent:         defaultPercent,
		}); err != nil {
			return nil, err
		}

		return &[]v1.Traffic{*t}, nil
	}

	var traffics []v1.Traffic
	trafficsModel := make([]apprunSharedTrafficsModel, 0, len(model.Traffics.Elements()))
	_ = model.Traffics.ElementsAs(context.Background(), &trafficsModel, false)
	for _, traffic := range trafficsModel {
		percent := int(traffic.Percent.ValueInt32())
		version_index := int(traffic.VersionIndex.ValueInt64())
		if len(versions) <= version_index {
			return nil, fmt.Errorf("index out of range, version_index: %d", version_index)
		}

		version := (versions)[version_index]
		tr := &v1.Traffic{}
		if err := tr.FromTrafficWithVersionName(v1.TrafficWithVersionName{
			Percent:     percent,
			VersionName: version.Name,
		}); err != nil {
			return nil, err
		}
		traffics = append(traffics, *tr)
	}

	return &traffics, nil
}

func expandApprunPacketFilter(model *apprunSharedResourceModel) *v1.PatchPacketFilter {
	enabled := false
	ret := &v1.PatchPacketFilter{
		IsEnabled: &enabled,
	}
	if !model.PacketFilter.IsNull() && !model.PacketFilter.IsUnknown() {
		var d apprunSharedPacketFilterModel
		_ = model.PacketFilter.As(context.Background(), &d, basetypes.ObjectAsOptions{})

		enabled = d.Enabled.ValueBool()
		var settings []v1.PacketFilterSetting
		for _, setting := range d.Settings {
			settings = append(settings, v1.PacketFilterSetting{
				FromIp:             setting.FromIP.ValueString(),
				FromIpPrefixLength: int(setting.FromIPPrefixLength.ValueInt32()),
			})
		}

		ret.IsEnabled = &enabled
		ret.Settings = &settings
	}
	return ret
}
