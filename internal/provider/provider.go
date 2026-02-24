// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package sakura

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sacloud/packages-go/envvar"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	"github.com/sacloud/terraform-provider-sakura/internal/service/apigw"
	"github.com/sacloud/terraform-provider-sakura/internal/service/apprun_shared"
	"github.com/sacloud/terraform-provider-sakura/internal/service/archive"
	"github.com/sacloud/terraform-provider-sakura/internal/service/auto_backup"
	"github.com/sacloud/terraform-provider-sakura/internal/service/auto_scale"
	"github.com/sacloud/terraform-provider-sakura/internal/service/bridge"
	"github.com/sacloud/terraform-provider-sakura/internal/service/cloudhsm"
	"github.com/sacloud/terraform-provider-sakura/internal/service/container_registry"
	"github.com/sacloud/terraform-provider-sakura/internal/service/database"
	"github.com/sacloud/terraform-provider-sakura/internal/service/dedicated_storage"
	"github.com/sacloud/terraform-provider-sakura/internal/service/disk"
	"github.com/sacloud/terraform-provider-sakura/internal/service/dns"
	"github.com/sacloud/terraform-provider-sakura/internal/service/dsr_lb"
	"github.com/sacloud/terraform-provider-sakura/internal/service/enhanced_db"
	"github.com/sacloud/terraform-provider-sakura/internal/service/enhanced_lb"
	"github.com/sacloud/terraform-provider-sakura/internal/service/eventbus"
	"github.com/sacloud/terraform-provider-sakura/internal/service/gslb"
	"github.com/sacloud/terraform-provider-sakura/internal/service/icon"
	"github.com/sacloud/terraform-provider-sakura/internal/service/internet"
	"github.com/sacloud/terraform-provider-sakura/internal/service/ipv4_ptr"
	"github.com/sacloud/terraform-provider-sakura/internal/service/kms"
	"github.com/sacloud/terraform-provider-sakura/internal/service/local_router"
	"github.com/sacloud/terraform-provider-sakura/internal/service/nfs"
	"github.com/sacloud/terraform-provider-sakura/internal/service/nosql"
	"github.com/sacloud/terraform-provider-sakura/internal/service/object_storage"
	"github.com/sacloud/terraform-provider-sakura/internal/service/packet_filter"
	"github.com/sacloud/terraform-provider-sakura/internal/service/private_host"
	secret_manager "github.com/sacloud/terraform-provider-sakura/internal/service/s3cret_manager"
	"github.com/sacloud/terraform-provider-sakura/internal/service/script"
	"github.com/sacloud/terraform-provider-sakura/internal/service/security_control"
	"github.com/sacloud/terraform-provider-sakura/internal/service/server"
	"github.com/sacloud/terraform-provider-sakura/internal/service/simple_monitor"
	"github.com/sacloud/terraform-provider-sakura/internal/service/simple_mq"
	"github.com/sacloud/terraform-provider-sakura/internal/service/simple_notification"
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
				Description: "The API token of your SakuraCloud account. It must be provided, but it can also be sourced from the `SAKURA_ACCESS_TOKEN`/`SAKURACLOUD_ACCESS_TOKEN` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"secret": schema.StringAttribute{
				Optional:    true,
				Description: "The API secret of your SakuraCloud account. It must be provided, but it can also be sourced from the `SAKURA_ACCESS_TOKEN_SECRET`/`SAKURACLOUD_ACCESS_TOKEN_SECRET` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"zone": schema.StringAttribute{
				Optional:    true,
				Description: "The name of zone to use as default. It must be provided, but it can also be sourced from the `SAKURA_ZONE`/`SAKURACLOUD_ZONE` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"zones": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of available SakuraCloud zone name. It can also be sourced via a shared credentials file if `profile` is specified. Default:[`is1a`, `is1b`, `tk1a`, `tk1v`]",
			},
			"default_zone": schema.StringAttribute{
				Optional:    true,
				Description: "The name of zone to use as default for global resources. It must be provided, but it can also be sourced from the `SAKURA_DEFAULT_ZONE`/`SAKURACLOUD_DEFAULT_ZONE` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"api_root_url": schema.StringAttribute{
				Optional:    true,
				Description: "The root URL of SakuraCloud API. It can also be sourced from the `SAKURA_API_ROOT_URL`/`SAKURACLOUD_API_ROOT_URL` environment variables, or via a shared credentials file if `profile` is specified. Default:`https://secure.sakura.ad.jp/cloud/zone`",
			},
			"retry_max": schema.Int32Attribute{
				Optional:    true,
				Description: "The maximum number of API call retries used when SakuraCloud API returns status code `423` or `503`. It can also be sourced from the `SAKURA_RETRY_MAX`/`SAKURACLOUD_RETRY_MAX` environment variables, or via a shared credentials file if `profile` is specified. Default:`100`",
				Validators: []validator.Int32{
					int32validator.Between(0, 100),
				},
			},
			"retry_wait_max": schema.Int64Attribute{
				Optional:    true,
				Description: "The maximum wait interval(in seconds) for retrying API call used when SakuraCloud API returns status code `423` or `503`.  It can also be sourced from the `SAKURA_RETRY_WAIT_MAX`/`SAKURACLOUD_RETRY_WAIT_MAX` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"retry_wait_min": schema.Int64Attribute{
				Optional:    true,
				Description: "The minimum wait interval(in seconds) for retrying API call used when SakuraCloud API returns status code `423` or `503`. It can also be sourced from the `SAKURA_RETRY_WAIT_MIN`/`SAKURACLOUD_RETRY_WAIT_MIN` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"api_request_timeout": schema.Int64Attribute{
				Optional: true,
				Description: desc.Sprintf(
					"The timeout seconds for each SakuraCloud API call. It can also be sourced from the `SAKURA_API_REQUEST_TIMEOUT`/`SAKURACLOUD_API_REQUEST_TIMEOUT` environment variables, or via a shared credentials file if `profile` is specified. Default:`%d`",
					common.APIRequestTimeout,
				),
			},
			"api_request_rate_limit": schema.Int32Attribute{
				Optional: true,
				Description: desc.Sprintf(
					"The maximum number of SakuraCloud API calls per second. It can also be sourced from the `SAKURA_RATE_LIMIT`/`SAKURACLOUD_RATE_LIMIT` environment variables, or via a shared credentials file if `profile` is specified. Default:`%d`",
					common.APIRequestRateLimit,
				),
				Validators: []validator.Int32{
					int32validator.Between(1, 10),
				},
			},
			"trace": schema.StringAttribute{
				Optional:    true,
				Description: "The flag to enable output trace log. It can also be sourced from the `SAKURA_TRACE`/`SAKURACLOUD_TRACE` environment variables, or via a shared credentials file if `profile` is specified",
			},
		},
	}
}

func (p *sakuraProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	envConf := &common.Config{
		Profile:             envvar.StringFromEnvMulti([]string{"SAKURA_PROFILE", "SAKURACLOUD_PROFILE"}, ""),
		AccessToken:         envvar.StringFromEnvMulti([]string{"SAKURA_ACCESS_TOKEN", "SAKURACLOUD_ACCESS_TOKEN"}, ""),
		AccessTokenSecret:   envvar.StringFromEnvMulti([]string{"SAKURA_ACCESS_TOKEN_SECRET", "SAKURACLOUD_ACCESS_TOKEN_SECRET"}, ""),
		Zone:                envvar.StringFromEnvMulti([]string{"SAKURA_ZONE", "SAKURACLOUD_ZONE"}, ""),
		DefaultZone:         envvar.StringFromEnvMulti([]string{"SAKURA_DEFAULT_ZONE", "SAKURACLOUD_DEFAULT_ZONE"}, ""),
		APIRootURL:          envvar.StringFromEnvMulti([]string{"SAKURA_API_ROOT_URL", "SAKURACLOUD_API_ROOT_URL"}, ""),
		RetryMax:            envvar.IntFromEnvMulti([]string{"SAKURA_RETRY_MAX", "SAKURACLOUD_RETRY_MAX"}, 0),
		RetryWaitMax:        envvar.IntFromEnvMulti([]string{"SAKURA_RETRY_WAIT_MAX", "SAKURACLOUD_RETRY_WAIT_MAX"}, 0),
		RetryWaitMin:        envvar.IntFromEnvMulti([]string{"SAKURA_RETRY_WAIT_MIN", "SAKURACLOUD_RETRY_WAIT_MIN"}, 0),
		APIRequestTimeout:   envvar.IntFromEnvMulti([]string{"SAKURA_API_REQUEST_TIMEOUT", "SAKURACLOUD_API_REQUEST_TIMEOUT"}, 0),
		APIRequestRateLimit: envvar.IntFromEnvMulti([]string{"SAKURA_RATE_LIMIT", "SAKURACLOUD_RATE_LIMIT"}, 0),
		Zones:               envvar.StringSliceFromEnvMulti([]string{"SAKURA_ZONES", "SAKURACLOUD_ZONES"}, nil),
	}

	var config sakuraProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.Config{
		Profile:             config.Profile.ValueString(),
		AccessToken:         config.AccessToken.ValueString(),
		AccessTokenSecret:   config.AccessTokenSecret.ValueString(),
		Zone:                config.Zone.ValueString(),
		DefaultZone:         config.DefaultZone.ValueString(),
		APIRootURL:          config.APIRootURL.ValueString(),
		RetryMax:            int(config.RetryMax.ValueInt32()),
		RetryWaitMax:        int(config.RetryWaitMax.ValueInt64()),
		RetryWaitMin:        int(config.RetryWaitMin.ValueInt64()),
		APIRequestTimeout:   int(config.APIRequestTimeout.ValueInt64()),
		APIRequestRateLimit: int(config.APIRequestRateLimit.ValueInt32()),
		TraceMode:           config.TraceMode.ValueString(),
		Zones:               common.TlistToStrings(config.Zones),
		TerraformVersion:    req.TerraformVersion,
	}
	// 他のパラメータとは違いプロファイルをロードするために、SAKURA_PROFILEの値だけは優先する
	if cfg.Profile == "" {
		cfg.Profile = envConf.Profile
	}

	var theClient saclient.Client
	if err := theClient.SettingsFromTerraformProvider(&config); err != nil {
		resp.Diagnostics.AddError("Error creating Sakura client", err.Error())
		return
	}

	client, err := cfg.NewClient(envConf, &theClient)
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
		apigw.NewApigwCertDataSource,
		apigw.NewApigwDomainDataSource,
		apigw.NewApigwGroupDataSource,
		apigw.NewApigwPlanDataSource,
		apigw.NewApigwRouteDataSource,
		apigw.NewApigwServiceDataSource,
		apigw.NewApigwSubscriptionDataSource,
		apigw.NewApigwUserDataSource,
		archive.NewArchiveDataSource,
		auto_scale.NewAutoScaleDataSource,
		bridge.NewBridgeDataSource,
		cloudhsm.NewCloudHSMDataSource,
		cloudhsm.NewCloudHSMClientDataSource,
		cloudhsm.NewCloudHSMPeerDataSource,
		cloudhsm.NewCloudHSMLicenseDataSource,
		container_registry.NewContainerRegistryDataSource,
		database.NewDatabaseDataSource,
		dedicated_storage.NewDedicatedStorageDataSource,
		disk.NewDiskDataSource,
		dns.NewDNSDataSource,
		dsr_lb.NewDSRLBDataSource,
		enhanced_db.NewEnhancedDBDataSource,
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
		security_control.NewActivationDataSource,
		security_control.NewAutomatedActionDataSource,
		security_control.NewEvaluationRuleDataSource,
		server.NewServerDataSource,
		simple_monitor.NewSimpleMonitorDataSource,
		simple_mq.NewSimpleMQDataSource,
		simple_notification.NewDestinationDataSource,
		simple_notification.NewGroupDataSource,
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
		apigw.NewApigwCertResource,
		apigw.NewApigwDomainResource,
		apigw.NewApigwGroupResource,
		apigw.NewApigwRouteResource,
		apigw.NewApigwServiceResource,
		apigw.NewApigwSubscriptionResource,
		apigw.NewApigwUserResource,
		archive.NewArchiveResource,
		auto_backup.NewAutoBackupResource,
		auto_scale.NewAutoScaleResource,
		bridge.NewBridgeResource,
		cloudhsm.NewCloudHSMResource,
		cloudhsm.NewCloudHSMClientResource,
		cloudhsm.NewCloudHSMPeerResource,
		cloudhsm.NewCloudHSMLicenseResource,
		container_registry.NewContainerRegistryResource,
		database.NewDatabaseResource,
		database.NewDatabaseReadReplicaResource,
		dedicated_storage.NewDedicatedStorageResource,
		disk.NewDiskResource,
		dns.NewDNSResource,
		dns.NewDNSRecordResource,
		dsr_lb.NewDSRLBResource,
		enhanced_db.NewEnhancedDBResource,
		enhanced_lb.NewEnhancedLBResource,
		enhanced_lb.NewEnhancedLBACMEResource,
		eventbus.NewEventBusProcessConfigurationResource,
		eventbus.NewEventBusScheduleResource,
		eventbus.NewEventBusTriggerResource,
		gslb.NewGSLBResource,
		icon.NewIconResource,
		internet.NewInternetResource,
		ipv4_ptr.NewIPv4PtrResource,
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
		security_control.NewActivationResource,
		security_control.NewAutomatedActionResource,
		security_control.NewEvaluationRuleResource,
		server.NewServerResource,
		simple_monitor.NewSimpleMonitorResource,
		simple_mq.NewSimpleMQResource,
		simple_notification.NewDestinationResource,
		simple_notification.NewGroupResource,
		ssh_key.NewSSHKeyResource,
		subnet.NewSubnetResource,
		sw1tch.NewSwitchResource,
		vswitch.NewvSwitchResource,
		vpn_router.NewVPNRouterResource,
		// ...他のリソースも同様に追加...
	}
}

func (p *sakuraProviderModel) LookupClientConfigProfileName() (string, bool) {
	return p.Profile.ValueString(), !p.Profile.IsNull() && !p.Profile.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigPrivateKeyPath() (string, bool) {
	return "", false // :TODO: :TBW:
}

func (p *sakuraProviderModel) LookupClientConfigServicePrincipalID() (string, bool) {
	return "", false // :TODO: :TBW:
}

func (p *sakuraProviderModel) LookupClientConfigServicePrincipalKeyID() (string, bool) {
	return "", false // :TODO: :TBW:
}

func (p *sakuraProviderModel) LookupClientConfigAccessToken() (string, bool) {
	return p.AccessToken.ValueString(), !p.AccessToken.IsNull() && !p.AccessToken.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigAccessTokenSecret() (string, bool) {
	return p.AccessTokenSecret.ValueString(), !p.AccessTokenSecret.IsNull() && !p.AccessTokenSecret.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigZone() (string, bool) {
	return p.Zone.ValueString(), !p.Zone.IsNull() && !p.Zone.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigDefaultZone() (string, bool) {
	return p.DefaultZone.ValueString(), !p.DefaultZone.IsNull() && !p.DefaultZone.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigZones() ([]string, bool) {
	return common.TlistToStrings(p.Zones), !p.Zones.IsNull() && !p.Zones.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigRetryMax() (int64, bool) {
	return int64(p.RetryMax.ValueInt32()), !p.RetryMax.IsNull() && !p.RetryMax.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigRetryWaitMax() (int64, bool) {
	return p.RetryWaitMax.ValueInt64(), !p.RetryWaitMax.IsNull() && !p.RetryWaitMax.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigRetryWaitMin() (int64, bool) {
	return p.RetryWaitMin.ValueInt64(), !p.RetryWaitMin.IsNull() && !p.RetryWaitMin.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigAPIRootURL() (string, bool) {
	return p.APIRootURL.ValueString(), !p.APIRootURL.IsNull() && !p.APIRootURL.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigAPIRequestTimeout() (int64, bool) {
	return p.APIRequestTimeout.ValueInt64(), !p.APIRequestTimeout.IsNull() && !p.APIRequestTimeout.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigAPIRequestRateLimit() (int64, bool) {
	return int64(p.APIRequestRateLimit.ValueInt32()), !p.APIRequestRateLimit.IsNull() && !p.APIRequestRateLimit.IsUnknown()
}

func (p *sakuraProviderModel) LookupClientConfigTraceMode() (string, bool) {
	return p.TraceMode.ValueString(), !p.TraceMode.IsNull() && !p.TraceMode.IsUnknown()
}

var _ saclient.TerraformProviderInterface = (*sakuraProviderModel)(nil)
