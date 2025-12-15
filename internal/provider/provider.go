// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package sakura

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	apiprof "github.com/sacloud/api-client-go/profile"
	"github.com/sacloud/packages-go/envvar"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	"github.com/sacloud/terraform-provider-sakura/internal/service/apprun_shared"
	"github.com/sacloud/terraform-provider-sakura/internal/service/archive"
	"github.com/sacloud/terraform-provider-sakura/internal/service/bridge"
	"github.com/sacloud/terraform-provider-sakura/internal/service/cloudhsm"
	"github.com/sacloud/terraform-provider-sakura/internal/service/container_registry"
	"github.com/sacloud/terraform-provider-sakura/internal/service/database"
	"github.com/sacloud/terraform-provider-sakura/internal/service/disk"
	"github.com/sacloud/terraform-provider-sakura/internal/service/dns"
	"github.com/sacloud/terraform-provider-sakura/internal/service/enhanced_lb"
	"github.com/sacloud/terraform-provider-sakura/internal/service/eventbus"
	"github.com/sacloud/terraform-provider-sakura/internal/service/gslb"
	"github.com/sacloud/terraform-provider-sakura/internal/service/icon"
	"github.com/sacloud/terraform-provider-sakura/internal/service/internet"
	"github.com/sacloud/terraform-provider-sakura/internal/service/kms"
	"github.com/sacloud/terraform-provider-sakura/internal/service/local_router"
	"github.com/sacloud/terraform-provider-sakura/internal/service/nfs"
	"github.com/sacloud/terraform-provider-sakura/internal/service/nosql"
	"github.com/sacloud/terraform-provider-sakura/internal/service/object_storage"
	"github.com/sacloud/terraform-provider-sakura/internal/service/packet_filter"
	"github.com/sacloud/terraform-provider-sakura/internal/service/private_host"
	secret_manager "github.com/sacloud/terraform-provider-sakura/internal/service/s3cret_manager"
	"github.com/sacloud/terraform-provider-sakura/internal/service/script"
	"github.com/sacloud/terraform-provider-sakura/internal/service/server"
	"github.com/sacloud/terraform-provider-sakura/internal/service/simple_monitor"
	"github.com/sacloud/terraform-provider-sakura/internal/service/simple_mq"
	"github.com/sacloud/terraform-provider-sakura/internal/service/ssh_key"
	"github.com/sacloud/terraform-provider-sakura/internal/service/subnet"
	sw1tch "github.com/sacloud/terraform-provider-sakura/internal/service/switch"
	"github.com/sacloud/terraform-provider-sakura/internal/service/vpn_router"
	"github.com/sacloud/terraform-provider-sakura/internal/service/vswitch"
	"github.com/sacloud/terraform-provider-sakura/internal/service/zone"
)

type sakuraProviderModel struct {
	Profile             types.String `tfsdk:"profile"`
	AccessToken         types.String `tfsdk:"token"`
	AccessTokenSecret   types.String `tfsdk:"secret"`
	Zone                types.String `tfsdk:"zone"`
	Zones               types.List   `tfsdk:"zones"`
	DefaultZone         types.String `tfsdk:"default_zone"`
	APIRootURL          types.String `tfsdk:"api_root_url"`
	RetryMax            types.Int32  `tfsdk:"retry_max"`
	RetryWaitMax        types.Int64  `tfsdk:"retry_wait_max"`
	RetryWaitMin        types.Int64  `tfsdk:"retry_wait_min"`
	APIRequestTimeout   types.Int64  `tfsdk:"api_request_timeout"`
	APIRequestRateLimit types.Int32  `tfsdk:"api_request_rate_limit"`
	TraceMode           types.String `tfsdk:"trace"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &sakuraProvider{version: version}
	}
}

type sakuraProvider struct {
	version string
	client  *common.APIClient
}

func (p *sakuraProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sakura"
	resp.Version = p.version
}

func (p *sakuraProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"profile": schema.StringAttribute{
				Optional:    true,
				Description: "The profile name of your SakuraCloud account. Default:`default`",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Description: "The API token of your SakuraCloud account. It must be provided, but it can also be sourced from the `SAKURACLOUD_ACCESS_TOKEN` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"secret": schema.StringAttribute{
				Optional:    true,
				Description: "The API secret of your SakuraCloud account. It must be provided, but it can also be sourced from the `SAKURACLOUD_ACCESS_TOKEN_SECRET` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Description: "The name of zone to use as default. It must be provided, but it can also be sourced from the `SAKURACLOUD_ZONE` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"zones": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of available SakuraCloud zone name. It can also be sourced via a shared credentials file if `profile` is specified. Default:[`is1a`, `is1b`, `tk1a`, `tk1v`]",
			},
			"default_zone": schema.StringAttribute{
				Optional:    true,
				Description: "The name of zone to use as default for global resources. It must be provided, but it can also be sourced from the `SAKURACLOUD_DEFAULT_ZONE` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"api_root_url": schema.StringAttribute{
				Optional:    true,
				Description: "The root URL of SakuraCloud API. It can also be sourced from the `SAKURACLOUD_API_ROOT_URL` environment variables, or via a shared credentials file if `profile` is specified. Default:`https://secure.sakura.ad.jp/cloud/zone`",
			},
			"retry_max": schema.Int32Attribute{
				Optional:    true,
				Description: "The maximum number of API call retries used when SakuraCloud API returns status code `423` or `503`. It can also be sourced from the `SAKURACLOUD_RETRY_MAX` environment variables, or via a shared credentials file if `profile` is specified. Default:`100`",
				Validators: []validator.Int32{
					int32validator.Between(0, 100),
				},
			},
			"retry_wait_max": schema.Int64Attribute{
				Optional:    true,
				Description: "The maximum wait interval(in seconds) for retrying API call used when SakuraCloud API returns status code `423` or `503`.  It can also be sourced from the `SAKURACLOUD_RETRY_WAIT_MAX` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"retry_wait_min": schema.Int64Attribute{
				Optional:    true,
				Description: "The minimum wait interval(in seconds) for retrying API call used when SakuraCloud API returns status code `423` or `503`. It can also be sourced from the `SAKURACLOUD_RETRY_WAIT_MIN` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"api_request_timeout": schema.Int64Attribute{
				Optional: true,
				Description: desc.Sprintf(
					"The timeout seconds for each SakuraCloud API call. It can also be sourced from the `SAKURACLOUD_API_REQUEST_TIMEOUT` environment variables, or via a shared credentials file if `profile` is specified. Default:`%d`",
					common.APIRequestTimeout,
				),
			},
			"api_request_rate_limit": schema.Int32Attribute{
				Optional: true,
				Description: desc.Sprintf(
					"The maximum number of SakuraCloud API calls per second. It can also be sourced from the `SAKURACLOUD_RATE_LIMIT` environment variables, or via a shared credentials file if `profile` is specified. Default:`%d`",
					common.APIRequestRateLimit,
				),
				Validators: []validator.Int32{
					int32validator.Between(1, 10),
				},
			},
			"trace": schema.StringAttribute{
				Optional:    true,
				Description: "The flag to enable output trace log. It can also be sourced from the `SAKURACLOUD_TRACE` environment variables, or via a shared credentials file if `profile` is specified",
			},
		},
	}
}

func getStrValueFromEnv(envVar string, defaultValue string) string {
	value, ok := os.LookupEnv(envVar)
	if !ok {
		return defaultValue
	}
	return value
}

func getIntValueFromEnv(resp *provider.ConfigureResponse, envVar string, defaultValue int) int {
	valueStr, ok := os.LookupEnv(envVar)
	if !ok {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		resp.Diagnostics.AddError("Environment Variable Error", fmt.Sprintf("failed to parse environment variable[%q]: %s", envVar, err.Error()))
		return defaultValue
	}
	return value
}

func (p *sakuraProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	profile := getStrValueFromEnv("SAKURACLOUD_PROFILE", apiprof.DefaultProfileName)
	token := getStrValueFromEnv("SAKURACLOUD_ACCESS_TOKEN", "")
	secret := getStrValueFromEnv("SAKURACLOUD_ACCESS_TOKEN_SECRET", "")
	zone := getStrValueFromEnv("SAKURACLOUD_ZONE", common.Zone)
	defaultZone := getStrValueFromEnv("SAKURACLOUD_DEFAULT_ZONE", "")
	apiRootUrl := getStrValueFromEnv("SAKURACLOUD_API_ROOT_URL", "")
	retryMax := getIntValueFromEnv(resp, "SAKURACLOUD_RETRY_MAX", common.RetryMax)
	retryWaitMax := getIntValueFromEnv(resp, "SAKURACLOUD_RETRY_WAIT_MAX", 0)
	retryWaitMin := getIntValueFromEnv(resp, "SAKURACLOUD_RETRY_WAIT_MIN", 0)
	apiRequestTimeout := getIntValueFromEnv(resp, "SAKURACLOUD_API_REQUEST_TIMEOUT", common.APIRequestTimeout)
	apiRequestRateLimit := getIntValueFromEnv(resp, "SAKURACLOUD_RATE_LIMIT", common.APIRequestRateLimit)

	var config sakuraProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Plugin Frameworkの設定値が最優先
	if config.Profile.ValueString() != "" {
		profile = config.Profile.ValueString()
	}
	if config.AccessToken.ValueString() != "" {
		token = config.AccessToken.ValueString()
	}
	if config.AccessTokenSecret.ValueString() != "" {
		secret = config.AccessTokenSecret.ValueString()
	}
	if config.Zone.ValueString() != "" {
		zone = config.Zone.ValueString()
	}
	if config.DefaultZone.ValueString() != "" {
		defaultZone = config.DefaultZone.ValueString()
	}
	if config.APIRootURL.ValueString() != "" {
		apiRootUrl = config.APIRootURL.ValueString()
	}
	if !config.RetryMax.IsNull() && !config.RetryMax.IsUnknown() {
		retryMax = int(config.RetryMax.ValueInt32())
	}
	if !config.RetryWaitMax.IsNull() && !config.RetryWaitMax.IsUnknown() {
		retryWaitMax = int(config.RetryWaitMax.ValueInt64())
	}
	if !config.RetryWaitMin.IsNull() && !config.RetryWaitMin.IsUnknown() {
		retryWaitMin = int(config.RetryWaitMin.ValueInt64())
	}
	if !config.APIRequestTimeout.IsNull() && !config.APIRequestTimeout.IsUnknown() {
		apiRequestTimeout = int(config.APIRequestTimeout.ValueInt64())
	}
	if !config.APIRequestRateLimit.IsNull() && !config.APIRequestRateLimit.IsUnknown() {
		apiRequestRateLimit = int(config.APIRequestRateLimit.ValueInt32())
	}
	zones := []string{}
	if !config.Zones.IsNull() && !config.Zones.IsUnknown() {
		for _, v := range config.Zones.Elements() {
			zones = append(zones, v.(types.String).ValueString())
		}
	}
	if len(zones) == 0 {
		zones = envvar.StringSliceFromEnv("SAKURACLOUD_ZONES", nil)
	}

	cfg := common.Config{
		Profile:             profile,
		AccessToken:         token,
		AccessTokenSecret:   secret,
		Zone:                zone,
		Zones:               zones,
		DefaultZone:         defaultZone,
		TraceMode:           config.TraceMode.ValueString(),
		APIRootURL:          apiRootUrl,
		RetryMax:            retryMax,
		RetryWaitMax:        retryWaitMax,
		RetryWaitMin:        retryWaitMin,
		APIRequestTimeout:   apiRequestTimeout,
		APIRequestRateLimit: apiRequestRateLimit,
		TerraformVersion:    req.TerraformVersion,
	}

	client, err := cfg.NewClient()
	if err != nil {
		resp.Diagnostics.AddError("Error creating Sakura client", err.Error())
		return
	}

	p.client = client
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *sakuraProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		apprun_shared.NewApprunSharedDataSource,
		archive.NewArchiveDataSource,
		bridge.NewBridgeDataSource,
		cloudhsm.NewCloudHSMDataSource,
		cloudhsm.NewCloudHSMClientDataSource,
		cloudhsm.NewCloudHSMPeerDataSource,
		cloudhsm.NewCloudHSMLicenseDataSource,
		container_registry.NewContainerRegistryDataSource,
		database.NewDatabaseDataSource,
		disk.NewDiskDataSource,
		dns.NewDNSDataSource,
		enhanced_lb.NewEnhancedLBDataSource,
		eventbus.NewEventBusProcessConfigurationDataSource,
		eventbus.NewEventBusScheduleDataSource,
		eventbus.NewEventBusTriggerDataSource,
		gslb.NewGSLBDataSource,
		icon.NewIconDataSource,
		internet.NewInternetDataSource,
		kms.NewKmsDataSource,
		local_router.NewLocalRouterDataSource,
		nfs.NewNFSDataSource,
		nosql.NewNosqlDataSource,
		object_storage.NewObjectStorageSiteDataSource,
		object_storage.NewObjectStorageBucketDataSource,
		object_storage.NewObjectStorageObjectDataSource,
		packet_filter.NewPacketFilterDataSource,
		private_host.NewPrivateHostDataSource,
		script.NewScriptDataSource,
		secret_manager.NewSecretManagerDataSource,
		secret_manager.NewSecretManagerSecretDataSource,
		server.NewServerDataSource,
		simple_monitor.NewSimpleMonitorDataSource,
		simple_mq.NewSimpleMQDataSource,
		ssh_key.NewSSHKeyDataSource,
		subnet.NewSubnetDataSource,
		sw1tch.NewSwitchDataSource,
		vswitch.NewvSwitchDataSource,
		vpn_router.NewVPNRouterDataSource,
		zone.NewZoneDataSource,
		// ...他のデータソースも同様に追加...
	}
}

func (p *sakuraProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		apprun_shared.NewApprunSharedResource,
		archive.NewArchiveResource,
		bridge.NewBridgeResource,
		cloudhsm.NewCloudHSMResource,
		cloudhsm.NewCloudHSMClientResource,
		cloudhsm.NewCloudHSMPeerResource,
		cloudhsm.NewCloudHSMLicenseResource,
		container_registry.NewContainerRegistryResource,
		database.NewDatabaseResource,
		database.NewDatabaseReadReplicaResource,
		disk.NewDiskResource,
		dns.NewDNSResource,
		dns.NewDNSRecordResource,
		enhanced_lb.NewEnhancedLBResource,
		enhanced_lb.NewEnhancedLBACMEResource,
		eventbus.NewEventBusProcessConfigurationResource,
		eventbus.NewEventBusScheduleResource,
		eventbus.NewEventBusTriggerResource,
		gslb.NewGSLBResource,
		icon.NewIconResource,
		internet.NewInternetResource,
		kms.NewKMSResource,
		local_router.NewLocalRouterResource,
		nfs.NewNFSResource,
		nosql.NewNosqlResource,
		nosql.NewNosqlAdditionalNodesResource,
		object_storage.NewObjectStorageBucketResource,
		object_storage.NewObjectStorageBucketCorsResource,
		object_storage.NewObjectStorageBucketVersioningResource,
		object_storage.NewObjectStorageObjectResource,
		object_storage.NewObjectStoragePermissionResource,
		packet_filter.NewPacketFilterResource,
		packet_filter.NewPacketFilterRulesResource,
		private_host.NewPrivateHostResource,
		script.NewScriptResource,
		secret_manager.NewSecretManagerResource,
		secret_manager.NewSecretManagerSecretResource,
		server.NewServerResource,
		simple_monitor.NewSimpleMonitorResource,
		simple_mq.NewSimpleMQResource,
		ssh_key.NewSSHKeyResource,
		subnet.NewSubnetResource,
		sw1tch.NewSwitchResource,
		vswitch.NewvSwitchResource,
		vpn_router.NewVPNRouterResource,
		// ...他のリソースも同様に追加...
	}
}
