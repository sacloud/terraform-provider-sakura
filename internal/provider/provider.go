// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package sakura

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/common/utils"
	"github.com/sacloud/terraform-provider-sakura/internal/desc"
	"github.com/sacloud/terraform-provider-sakura/internal/service/addon"
	"github.com/sacloud/terraform-provider-sakura/internal/service/apigw"
	"github.com/sacloud/terraform-provider-sakura/internal/service/apprun_dedicated"
	"github.com/sacloud/terraform-provider-sakura/internal/service/apprun_shared"
	"github.com/sacloud/terraform-provider-sakura/internal/service/archive"
	"github.com/sacloud/terraform-provider-sakura/internal/service/auto_backup"
	"github.com/sacloud/terraform-provider-sakura/internal/service/auto_scale"
	"github.com/sacloud/terraform-provider-sakura/internal/service/bridge"
	"github.com/sacloud/terraform-provider-sakura/internal/service/cdrom"
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
	"github.com/sacloud/terraform-provider-sakura/internal/service/iam"
	"github.com/sacloud/terraform-provider-sakura/internal/service/icon"
	"github.com/sacloud/terraform-provider-sakura/internal/service/internet"
	"github.com/sacloud/terraform-provider-sakura/internal/service/ipv4_ptr"
	"github.com/sacloud/terraform-provider-sakura/internal/service/kms"
	"github.com/sacloud/terraform-provider-sakura/internal/service/local_router"
	"github.com/sacloud/terraform-provider-sakura/internal/service/monitoring_suite"
	"github.com/sacloud/terraform-provider-sakura/internal/service/networking_suite"
	"github.com/sacloud/terraform-provider-sakura/internal/service/nfs"
	"github.com/sacloud/terraform-provider-sakura/internal/service/nosql"
	"github.com/sacloud/terraform-provider-sakura/internal/service/object_storage"
	"github.com/sacloud/terraform-provider-sakura/internal/service/ondemand_db"
	"github.com/sacloud/terraform-provider-sakura/internal/service/packet_filter"
	"github.com/sacloud/terraform-provider-sakura/internal/service/private_host"
	secret_manager "github.com/sacloud/terraform-provider-sakura/internal/service/s3cret_manager"
	"github.com/sacloud/terraform-provider-sakura/internal/service/script"
	"github.com/sacloud/terraform-provider-sakura/internal/service/security_control"
	service_endpoint_gateway "github.com/sacloud/terraform-provider-sakura/internal/service/seg"
	"github.com/sacloud/terraform-provider-sakura/internal/service/server"
	"github.com/sacloud/terraform-provider-sakura/internal/service/simple_monitor"
	"github.com/sacloud/terraform-provider-sakura/internal/service/simple_mq"
	"github.com/sacloud/terraform-provider-sakura/internal/service/simple_notification"
	"github.com/sacloud/terraform-provider-sakura/internal/service/ssh_key"
	"github.com/sacloud/terraform-provider-sakura/internal/service/subnet"
	sw1tch "github.com/sacloud/terraform-provider-sakura/internal/service/switch"
	"github.com/sacloud/terraform-provider-sakura/internal/service/vpn_router"
	"github.com/sacloud/terraform-provider-sakura/internal/service/vswitch"
	"github.com/sacloud/terraform-provider-sakura/internal/service/webaccel"
	"github.com/sacloud/terraform-provider-sakura/internal/service/workflows"
	"github.com/sacloud/terraform-provider-sakura/internal/service/zone"
)

type sakuraProviderModel struct {
	Profile                types.String `tfsdk:"profile"`
	AccessToken            types.String `tfsdk:"token"`
	AccessTokenSecret      types.String `tfsdk:"secret"`
	ServicePrincipalID     types.String `tfsdk:"service_principal_id"`
	ServicePrincipalKeyID  types.String `tfsdk:"service_principal_key_id"`
	ServicePrincipalKeyKID types.String `tfsdk:"service_principal_key_kid"`
	ServicePrivateKeyPath  types.String `tfsdk:"service_private_key_path"`
	Zone                   types.String `tfsdk:"zone"`
	Zones                  types.List   `tfsdk:"zones"`
	DefaultZone            types.String `tfsdk:"default_zone"`
	APIRootURL             types.String `tfsdk:"api_root_url"`
	RetryMax               types.Int32  `tfsdk:"retry_max"`
	RetryWaitMax           types.Int64  `tfsdk:"retry_wait_max"`
	RetryWaitMin           types.Int64  `tfsdk:"retry_wait_min"`
	APIRequestTimeout      types.Int64  `tfsdk:"api_request_timeout"`
	APIRequestRateLimit    types.Int32  `tfsdk:"api_request_rate_limit"`
	TraceMode              types.String `tfsdk:"trace"`
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
			"service_principal_id": schema.StringAttribute{
				Optional:    true,
				Description: "The service principal ID of your SakuraCloud account. This is used for service principal based access. It can also be sourced from the `SAKURA_SERVICE_PRINCIPAL_ID` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"service_principal_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "Deprecated: Use `service_principal_key_kid` instead. The service principal key ID of your SakuraCloud account. This is used for service principal based access. It can also be sourced from the `SAKURA_SERVICE_PRINCIPAL_KEY_ID` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"service_principal_key_kid": schema.StringAttribute{
				Optional:    true,
				Description: "The service principal key KID of your SakuraCloud account. This is used for service principal based access. It can also be sourced from the `SAKURA_SERVICE_PRINCIPAL_KEY_KID` environment variables, or via a shared credentials file if `profile` is specified",
			},
			"service_private_key_path": schema.StringAttribute{
				Optional:    true,
				Description: "The private key path for service principal of your SakuraCloud account. This is used for service principal based access. It can also be sourced from the `SAKURA_PRIVATE_KEY_PATH` environment variables, or via a shared credentials file if `profile` is specified",
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
	var theClient saclient.Client
	var config sakuraProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if utils.IsKnown(config.ServicePrincipalKeyKID) {
	} else if utils.IsKnown(config.ServicePrincipalKeyID) {
		resp.Diagnostics.AddWarning("Deprecated Configuration", "The `service_principal_key_id` attribute is deprecated and will be removed in the future. Use `service_principal_key_kid` instead.")
	}

	err := theClient.SettingsFromTerraformProvider(&config)
	if err != nil {
		resp.Diagnostics.AddError("failed to set settings from Terraform provider", err.Error())
		return
	}

	err = theClient.SetWith(saclient.WithUserAgent(p.httpUserAgent(ctx, req)))
	if err != nil {
		resp.Diagnostics.AddError("failed to set user agent", err.Error())
		return
	}

	client, err := common.NewClient(&theClient)
	if err != nil {
		resp.Diagnostics.AddError("failed to create Sakura client", err.Error())
		return
	}

	p.client = client
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *sakuraProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		addon.NewAIDataSource,
		addon.NewCDNDataSource,
		addon.NewDataLakeDataSource,
		addon.NewDDoSDataSource,
		addon.NewDWHDataSource,
		addon.NewETLDataSource,
		addon.NewQueryDataSource,
		addon.NewSearchDataSource,
		addon.NewStreamingDataSource,
		addon.NewWAFDataSource,
		apigw.NewApigwCertDataSource,
		apigw.NewApigwDomainDataSource,
		apigw.NewApigwGroupDataSource,
		apigw.NewApigwPlanDataSource,
		apigw.NewApigwRouteDataSource,
		apigw.NewApigwServiceDataSource,
		apigw.NewApigwSubscriptionDataSource,
		apigw.NewApigwUserDataSource,
		apprun_dedicated.NewAppDataSource,
		apprun_dedicated.NewAutoScalingGroupDataSource,
		apprun_dedicated.NewCertDataSource,
		apprun_dedicated.NewClusterDataSource,
		apprun_dedicated.NewLoadBalancerDataSource,
		apprun_dedicated.NewLoadBalancerNodesDataSource,
		apprun_dedicated.NewLoadBalancersDataSource,
		apprun_dedicated.NewLoadBalancerServiceClassesDataSource,
		apprun_dedicated.NewVersionDataSource,
		apprun_dedicated.NewWorkerNodesDataSource,
		apprun_dedicated.NewWorkerServiceClassesDataSource,
		apprun_shared.NewApprunSharedDataSource,
		archive.NewArchiveDataSource,
		auto_scale.NewAutoScaleDataSource,
		bridge.NewBridgeDataSource,
		cdrom.NewCDROMDataSource,
		cloudhsm.NewCloudHSMClientDataSource,
		cloudhsm.NewCloudHSMDataSource,
		cloudhsm.NewCloudHSMLicenseDataSource,
		cloudhsm.NewCloudHSMPeerDataSource,
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
		iam.NewAuthDataSource,
		iam.NewAuthContextDataSource,
		iam.NewFolderDataSource,
		iam.NewGroupDataSource,
		iam.NewIdRoleDataSource,
		iam.NewOrgIDPolicyDataSource,
		iam.NewPolicyDataSource,
		iam.NewProjectApiKeyDataSource,
		iam.NewProjectDataSource,
		iam.NewRoleDataSource,
		iam.NewServicePrincipalDataSource,
		iam.NewSsoDataSource,
		iam.NewUserDataSource,
		iam.NewUserProvisioningDataSource,
		icon.NewIconDataSource,
		internet.NewInternetDataSource,
		kms.NewKmsDataSource,
		local_router.NewLocalRouterDataSource,
		monitoring_suite.NewAlertProjectDataSource,
		monitoring_suite.NewAlertLogMeasureRuleDataSource,
		monitoring_suite.NewAlertNotificationRoutingDataSource,
		monitoring_suite.NewAlertNotificationTargetDataSource,
		monitoring_suite.NewAlertRuleDataSource,
		monitoring_suite.NewDashboardDataSource,
		monitoring_suite.NewLogStorageDataSource,
		monitoring_suite.NewMetricStorageDataSource,
		monitoring_suite.NewTraceStorageDataSource,
		networking_suite.NewSubnetGroupDataSource,
		networking_suite.NewSubnetDataSource,
		nfs.NewNFSDataSource,
		nosql.NewNosqlDataSource,
		object_storage.NewObjectStorageBucketDataSource,
		object_storage.NewObjectStorageObjectDataSource,
		object_storage.NewObjectStorageSiteDataSource,
		ondemand_db.NewOnDemandDBDataSource,
		packet_filter.NewPacketFilterDataSource,
		private_host.NewPrivateHostDataSource,
		script.NewScriptDataSource,
		secret_manager.NewSecretManagerDataSource,
		secret_manager.NewSecretManagerSecretDataSource,
		security_control.NewActivationDataSource,
		security_control.NewAutomatedActionDataSource,
		security_control.NewEvaluationRuleDataSource,
		server.NewServerDataSource,
		service_endpoint_gateway.NewSEGDataSource,
		simple_monitor.NewSimpleMonitorDataSource,
		simple_mq.NewSimpleMQDataSource,
		simple_notification.NewDestinationDataSource,
		simple_notification.NewGroupDataSource,
		simple_notification.NewRoutingDataSource,
		ssh_key.NewSSHKeyDataSource,
		subnet.NewSubnetDataSource,
		sw1tch.NewSwitchDataSource,
		vpn_router.NewVPNRouterDataSource,
		vswitch.NewvSwitchDataSource,
		webaccel.NewWebAccelDataSource,
		workflows.NewPlanDataSource,
		workflows.NewSubscriptionDataSource,
		workflows.NewWorkflowsDataSource,
		workflows.NewWorkflowsRevisionAliasDataSource,
		zone.NewZoneDataSource,
		// ...他のデータソースも同様に追加...
	}
}

func (p *sakuraProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		addon.NewAIResource,
		addon.NewCDNResource,
		addon.NewDataLakeResource,
		addon.NewDDoSResource,
		addon.NewDWHResource,
		addon.NewETLResource,
		addon.NewQueryResource,
		addon.NewSearchResource,
		addon.NewStreamingResource,
		addon.NewWAFResource,
		apigw.NewApigwCertResource,
		apigw.NewApigwDomainResource,
		apigw.NewApigwGroupResource,
		apigw.NewApigwRouteResource,
		apigw.NewApigwServiceResource,
		apigw.NewApigwSubscriptionResource,
		apigw.NewApigwUserResource,
		apprun_dedicated.NewAppResource,
		apprun_dedicated.NewAutoScalingGroupResource,
		apprun_dedicated.NewCertResource,
		apprun_dedicated.NewClusterResource,
		apprun_dedicated.NewLoadBalancerResource,
		apprun_dedicated.NewVersionResource,
		apprun_shared.NewApprunSharedResource,
		archive.NewArchiveResource,
		auto_backup.NewAutoBackupResource,
		auto_scale.NewAutoScaleResource,
		bridge.NewBridgeResource,
		cdrom.NewCDROMResource,
		cloudhsm.NewCloudHSMClientResource,
		cloudhsm.NewCloudHSMLicenseResource,
		cloudhsm.NewCloudHSMPeerResource,
		cloudhsm.NewCloudHSMResource,
		container_registry.NewContainerRegistryResource,
		database.NewDatabaseReadReplicaResource,
		database.NewDatabaseResource,
		dedicated_storage.NewDedicatedStorageResource,
		disk.NewDiskResource,
		dns.NewDNSRecordResource,
		dns.NewDNSResource,
		dsr_lb.NewDSRLBResource,
		enhanced_db.NewEnhancedDBResource,
		enhanced_lb.NewEnhancedLBACMEResource,
		enhanced_lb.NewEnhancedLBResource,
		eventbus.NewEventBusProcessConfigurationResource,
		eventbus.NewEventBusScheduleResource,
		eventbus.NewEventBusTriggerResource,
		gslb.NewGSLBResource,
		iam.NewAuthResource,
		iam.NewFolderResource,
		iam.NewGroupResource,
		iam.NewOrgIDPolicyResource,
		iam.NewPolicyResource,
		iam.NewProjectApiKeyResource,
		iam.NewProjectResource,
		iam.NewServicePrincipalResource,
		iam.NewSsoResource,
		iam.NewUserResource,
		iam.NewUserProvisioningResource,
		icon.NewIconResource,
		internet.NewInternetResource,
		ipv4_ptr.NewIPv4PtrResource,
		kms.NewKMSResource,
		local_router.NewLocalRouterResource,
		monitoring_suite.NewAlertProjectResource,
		monitoring_suite.NewAlertLogMeasureRuleResource,
		monitoring_suite.NewAlertNotificationRoutingResource,
		monitoring_suite.NewAlertNotificationTargetResource,
		monitoring_suite.NewAlertRuleResource,
		monitoring_suite.NewDashboardResource,
		monitoring_suite.NewLogRoutingResource,
		monitoring_suite.NewLogStorageAccessKeyResource,
		monitoring_suite.NewLogStorageResource,
		monitoring_suite.NewMetricRoutingResource,
		monitoring_suite.NewMetricStorageAccessKeyResource,
		monitoring_suite.NewMetricStorageResource,
		monitoring_suite.NewTraceStorageAccessKeyResource,
		monitoring_suite.NewTraceStorageResource,
		networking_suite.NewSubnetGroupResource,
		networking_suite.NewSubnetResource,
		networking_suite.NewInterfaceConnectionResource,
		nfs.NewNFSResource,
		nosql.NewNosqlAdditionalNodesResource,
		nosql.NewNosqlResource,
		object_storage.NewObjectStorageBucketCorsResource,
		object_storage.NewObjectStorageBucketEncryptionConfigResource,
		object_storage.NewObjectStorageBucketReplicationConfigResource,
		object_storage.NewObjectStorageBucketResource,
		object_storage.NewObjectStorageBucketVersioningResource,
		object_storage.NewObjectStorageObjectResource,
		object_storage.NewObjectStoragePermissionResource,
		ondemand_db.NewOnDemandDBResource,
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
		service_endpoint_gateway.NewSEGResource,
		simple_monitor.NewSimpleMonitorResource,
		simple_mq.NewSimpleMQResource,
		simple_notification.NewDestinationResource,
		simple_notification.NewGroupResource,
		simple_notification.NewRoutingResource,
		ssh_key.NewSSHKeyResource,
		subnet.NewSubnetResource,
		sw1tch.NewSwitchResource,
		vpn_router.NewVPNRouterResource,
		vswitch.NewvSwitchResource,
		webaccel.NewWebAccelResource,
		webaccel.NewWebAccelACLResource,
		webaccel.NewWebAccelActivationResource,
		webaccel.NewWebAccelCertificateResource,
		workflows.NewSubscriptionResource,
		workflows.NewWorkflowsResource,
		workflows.NewWorkflowsRevisionAliasResource,
		// ...他のリソースも同様に追加...
	}
}

const (
	uaEnvVar   = "SAKURACLOUD_APPEND_USER_AGENT"
	tfUAEnvVar = "TF_APPEND_USER_AGENT"
)

func (p *sakuraProvider) httpUserAgent(ctx context.Context, req provider.ConfigureRequest) string {
	tfver := req.TerraformVersion
	providerVar := p.version
	ua := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io)", tfver)

	if add := os.Getenv(tfUAEnvVar); add != "" {
		add = strings.TrimSpace(add)
		if len(add) > 0 {
			ua += " " + add
			tflog.Debug(ctx, "Using modified User-Agent", map[string]interface{}{"user_agent": ua})
		}
	}

	providerUserAgent := fmt.Sprintf("%s/v%s", "terraform-provider-sakura", providerVar)
	if add := os.Getenv(uaEnvVar); add != "" {
		add = strings.TrimSpace(add)
		if len(add) > 0 {
			ua += " " + add
			tflog.Debug(ctx, "Using modified User-Agent", map[string]interface{}{"user_agent": ua})
		}
	}

	return fmt.Sprintf("%s %s", ua, providerUserAgent)
}

func lookup[T any, V attr.Value](v V, f func(V) T) (ret T, ok bool) {
	if ok = utils.IsKnown(v); ok {
		ret = f(v)
	}
	return
}

func (p *sakuraProviderModel) LookupClientConfigProfileName() (string, bool) {
	return lookup(p.Profile, types.String.ValueString)
}

func (p *sakuraProviderModel) LookupClientConfigPrivateKeyPath() (string, bool) {
	return lookup(p.ServicePrivateKeyPath, types.String.ValueString)
}

func (p *sakuraProviderModel) LookupClientConfigServicePrincipalID() (string, bool) {
	return lookup(p.ServicePrincipalID, types.String.ValueString)
}

func (p *sakuraProviderModel) LookupClientConfigServicePrincipalKeyKID() (string, bool) {
	if kid, ok := lookup(p.ServicePrincipalKeyKID, types.String.ValueString); ok {
		return kid, ok
	} else if keyid, ok := lookup(p.ServicePrincipalKeyID, types.String.ValueString); ok {
		// See above for issuing warnings
		return keyid, ok
	} else {
		return "", false
	}
}

func (p *sakuraProviderModel) LookupClientConfigAccessToken() (string, bool) {
	return lookup(p.AccessToken, types.String.ValueString)
}

func (p *sakuraProviderModel) LookupClientConfigAccessTokenSecret() (string, bool) {
	return lookup(p.AccessTokenSecret, types.String.ValueString)
}

func (p *sakuraProviderModel) LookupClientConfigZone() (string, bool) {
	return lookup(p.Zone, types.String.ValueString)
}

func (p *sakuraProviderModel) LookupClientConfigDefaultZone() (string, bool) {
	return lookup(p.DefaultZone, types.String.ValueString)
}

func (p *sakuraProviderModel) LookupClientConfigZones() ([]string, bool) {
	return lookup(p.Zones, common.TlistToStrings)
}

func (p *sakuraProviderModel) LookupClientConfigRetryMax() (int64, bool) {
	i32, ok := lookup(p.RetryMax, types.Int32.ValueInt32)
	return int64(i32), ok
}

func (p *sakuraProviderModel) LookupClientConfigRetryWaitMax() (int64, bool) {
	return lookup(p.RetryWaitMax, types.Int64.ValueInt64)
}

func (p *sakuraProviderModel) LookupClientConfigRetryWaitMin() (int64, bool) {
	return lookup(p.RetryWaitMin, types.Int64.ValueInt64)
}

func (p *sakuraProviderModel) LookupClientConfigAPIRootURL() (string, bool) {
	return lookup(p.APIRootURL, types.String.ValueString)
}

func (p *sakuraProviderModel) LookupClientConfigAPIRequestTimeout() (int64, bool) {
	return lookup(p.APIRequestTimeout, types.Int64.ValueInt64)
}

func (p *sakuraProviderModel) LookupClientConfigAPIRequestRateLimit() (int64, bool) {
	i32, ok := lookup(p.APIRequestRateLimit, types.Int32.ValueInt32)
	return int64(i32), ok
}

func (p *sakuraProviderModel) LookupClientConfigTraceMode() (string, bool) {
	return lookup(p.TraceMode, types.String.ValueString)
}

var _ saclient.TerraformProviderInterface = (*sakuraProviderModel)(nil)
