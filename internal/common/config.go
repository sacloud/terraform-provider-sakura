// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/apigw-api-go"
	apigwapi "github.com/sacloud/apigw-api-go/apis/v1"
	"github.com/sacloud/apprun-api-go"
	dedicatedstorage "github.com/sacloud/dedicated-storage-api-go"
	dedicatedstorageapi "github.com/sacloud/dedicated-storage-api-go/apis/v1"
	"github.com/sacloud/eventbus-api-go"
	eventbus_api "github.com/sacloud/eventbus-api-go/apis/v1"
	saht "github.com/sacloud/go-http"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/api"
	"github.com/sacloud/iaas-api-go/helper/query"
	"github.com/sacloud/iam-api-go"
	iamapi "github.com/sacloud/iam-api-go/apis/v1"
	kms "github.com/sacloud/kms-api-go"
	kmsapi "github.com/sacloud/kms-api-go/apis/v1"
	nosql "github.com/sacloud/nosql-api-go"
	nosqlapi "github.com/sacloud/nosql-api-go/apis/v1"
	objectstorage "github.com/sacloud/object-storage-api-go"
	"github.com/sacloud/saclient-go"
	sm "github.com/sacloud/secretmanager-api-go"
	smapi "github.com/sacloud/secretmanager-api-go/apis/v1"
	seccon "github.com/sacloud/security-control-api-go"
	secconapi "github.com/sacloud/security-control-api-go/apis/v1"
	simple_notification "github.com/sacloud/simple-notification-api-go"
	simple_notification_api "github.com/sacloud/simple-notification-api-go/apis/v1"
	"github.com/sacloud/simplemq-api-go"
	"github.com/sacloud/simplemq-api-go/apis/v1/queue"
	"github.com/sacloud/terraform-provider-sakura/internal/defaults"
	ver "github.com/sacloud/terraform-provider-sakura/version"
)

const (
	traceHTTP = "http"
	traceAPI  = "api"
)

const (
	uaEnvVar        = "SAKURACLOUD_APPEND_USER_AGENT"
	EndpointsEnvVar = "SAKURA_ENDPOINTS_"
)

const (
	Zone                = "is1b"
	RetryMax            = 10
	APIRequestTimeout   = 300
	APIRequestRateLimit = 10
)

var (
	deletionWaiterTimeout            = 30 * time.Minute
	deletionWaiterPollingInterval    = 5 * time.Second
	databaseWaitAfterCreateDuration  = 1 * time.Minute
	vpcRouterWaitAfterCreateDuration = 2 * time.Minute
)

// Config type of SakuraCloud Config
type Config struct {
	Profile               string
	AccessToken           string
	AccessTokenSecret     string
	ServicePrincipalID    string
	ServicePrincipalKeyID string
	ServicePrivateKey     string
	ServicePrivateKeyPath string
	Zone                  string
	Zones                 []string
	DefaultZone           string
	TraceMode             string
	AcceptLanguage        string
	APIRootURL            string
	RetryMax              int
	RetryWaitMin          int
	RetryWaitMax          int
	APIRequestTimeout     int
	APIRequestRateLimit   int
	TerraformVersion      string
	Endpoints             map[string]string
}

// APIClient for SakuraCloud API
type APIClient struct {
	iaas.APICaller
	defaultZone                      string // 各リソースでzone未指定の場合に利用するゾーン。iaas.APIDefaultZoneとは別物。
	zones                            []string
	deletionWaiterTimeout            time.Duration
	deletionWaiterPollingInterval    time.Duration
	databaseWaitAfterCreateDuration  time.Duration
	vpcRouterWaitAfterCreateDuration time.Duration
	CallerOptions                    *client.Options
	AppRunClient                     *apprun.Client
	KmsClient                        *kmsapi.Client
	SecretManagerClient              *smapi.Client
	SimpleMqClient                   *queue.Client
	EventBusClient                   *eventbus_api.Client
	ObjectStorageClient              *objectstorage.Client
	NosqlClient                      *nosqlapi.Client
	DedicatedStorageClient           *dedicatedstorageapi.Client
	ApigwClient                      *apigwapi.Client
	SecurityControlClient            *secconapi.Client
	IamClient                        *iamapi.Client
	SimpleNotificationClient         *simple_notification_api.Client
}

func (c *APIClient) CheckReferencedOption() query.CheckReferencedOption {
	return query.CheckReferencedOption{
		Tick:    c.deletionWaiterPollingInterval,
		Timeout: c.deletionWaiterTimeout,
	}
}

func (c *APIClient) GetZones() []string {
	return c.zones
}

func (c *Config) FillWith(other *Config) {
	if c.Profile == "" {
		c.Profile = other.Profile
	}
	if c.AccessToken == "" {
		c.AccessToken = other.AccessToken
	}
	if c.AccessTokenSecret == "" {
		c.AccessTokenSecret = other.AccessTokenSecret
	}
	if c.ServicePrincipalID == "" {
		c.ServicePrincipalID = other.ServicePrincipalID
	}
	if c.ServicePrincipalKeyID == "" {
		c.ServicePrincipalKeyID = other.ServicePrincipalKeyID
	}
	if c.ServicePrivateKey == "" {
		c.ServicePrivateKey = other.ServicePrivateKey
	}
	if c.ServicePrivateKeyPath == "" {
		c.ServicePrivateKeyPath = other.ServicePrivateKeyPath
	}
	if c.Zone == "" {
		c.Zone = other.Zone
	}
	if c.DefaultZone == "" {
		c.DefaultZone = other.DefaultZone
	}
	if len(c.Zones) == 0 {
		c.Zones = other.Zones
	}
	if c.APIRootURL == "" {
		c.APIRootURL = other.APIRootURL
	}
	if c.TraceMode == "" {
		c.TraceMode = other.TraceMode
	}
	if c.AcceptLanguage == "" {
		c.AcceptLanguage = other.AcceptLanguage
	}
	if c.RetryMax == 0 && other.RetryMax > 0 {
		c.RetryMax = other.RetryMax
	}
	if c.RetryWaitMax == 0 {
		c.RetryWaitMax = other.RetryWaitMax
	}
	if c.RetryWaitMin == 0 {
		c.RetryWaitMin = other.RetryWaitMin
	}
	if c.APIRequestTimeout == 0 && other.APIRequestTimeout > 0 {
		c.APIRequestTimeout = other.APIRequestTimeout
	}
	if c.APIRequestRateLimit == 0 && other.APIRequestRateLimit > 0 {
		c.APIRequestRateLimit = other.APIRequestRateLimit
	}
	if len(c.Endpoints) == 0 && len(other.Endpoints) > 0 {
		c.Endpoints = other.Endpoints
	}
}

func (c *Config) FillWithDefault() {
	if c.Zone == "" {
		c.Zone = defaults.Zone
	}
	if len(c.Zones) == 0 {
		c.Zones = iaas.SakuraCloudZones
	}
	if c.RetryMax == 0 {
		c.RetryMax = defaults.RetryMax
	}
	if c.APIRequestTimeout == 0 {
		c.APIRequestTimeout = defaults.APIRequestTimeout
	}
	if c.APIRequestRateLimit == 0 {
		c.APIRequestRateLimit = defaults.APIRequestRateLimit
	}
}

func (c *Config) LoadFromProfile() (*Config, error) {
	profileOp := saclient.NewProfileOp(os.Environ())
	var attrs map[string]any
	if c.Profile == "" {
		if name, err := profileOp.GetCurrentName(); err != nil {
			if profile, err := profileOp.Read(defaults.DefaultProfileName); err != nil {
				return &Config{}, nil
			} else {
				c.Profile = defaults.DefaultProfileName
				attrs = profile.Attributes
			}
		} else {
			if profile, err := profileOp.Read(name); err != nil {
				return nil, fmt.Errorf("failed to load profile[%s]: %s", name, err)
			} else {
				c.Profile = name
				attrs = profile.Attributes
			}
		}
	} else {
		if profile, err := profileOp.Read(c.Profile); err != nil {
			return nil, fmt.Errorf("failed to load profile[%s]: %s", c.Profile, err)
		} else {
			attrs = profile.Attributes
		}
	}

	conf := &Config{}
	if v, ok := attrs["AccessToken"].(string); ok {
		conf.AccessToken = v
	}
	if v, ok := attrs["AccessTokenSecret"].(string); ok {
		conf.AccessTokenSecret = v
	}
	if v, ok := attrs["ServicePrincipalID"].(string); ok {
		conf.ServicePrincipalID = v
	}
	if v, ok := attrs["ServicePrincipalKeyID"].(string); ok {
		conf.ServicePrincipalKeyID = v
	}
	if v, ok := attrs["PrivateKeyPEMPath"].(string); ok {
		conf.ServicePrivateKeyPath = v
	}
	if v, ok := attrs["Zone"].(string); ok {
		conf.Zone = v
	}
	if v, ok := attrs["Zones"].([]any); ok {
		zones := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("invalid type for Zones element: %T", item)
			}
			zones = append(zones, s)
		}
		conf.Zones = zones
	}
	if v, ok := attrs["DefaultZone"].(string); ok {
		conf.DefaultZone = v
	}
	if v, ok := attrs["TraceMode"].(string); ok {
		conf.TraceMode = v
	}
	if v, ok := attrs["AcceptLanguage"].(string); ok {
		conf.AcceptLanguage = v
	}
	if v, ok := attrs["APIRootURL"].(string); ok {
		conf.APIRootURL = v
	}
	if v, ok := attrs["RetryMax"].(float64); ok {
		conf.RetryMax = int(v)
	}
	if v, ok := attrs["RetryWaitMin"].(float64); ok {
		conf.RetryWaitMin = int(v)
	}
	if v, ok := attrs["RetryWaitMax"].(float64); ok {
		conf.RetryWaitMax = int(v)
	}
	if v, ok := attrs["HTTPRequestTimeout"].(float64); ok {
		conf.APIRequestTimeout = int(v)
	}
	if v, ok := attrs["HTTPRequestRateLimit"].(float64); ok {
		conf.APIRequestRateLimit = int(v)
	}
	if v, ok := attrs["Endpoints"].(map[string]string); ok {
		conf.Endpoints = v
	}

	return conf, nil
}

func (c *Config) validate() error {
	var err error
	if c.ServicePrivateKey != "" || c.ServicePrivateKeyPath != "" {
		if c.ServicePrincipalID == "" {
			err = multierror.Append(err, errors.New("service_principal_id is required when service_private_key or service_private_key_path is specified"))
		}
		if c.ServicePrincipalKeyID == "" {
			err = multierror.Append(err, errors.New("service_principal_key_id is required when service_private_key or service_private_key_path is specified"))
		}
	} else {
		if c.AccessToken == "" {
			err = multierror.Append(err, errors.New("token is required"))
		}
		if c.AccessTokenSecret == "" {
			err = multierror.Append(err, errors.New("secret is required"))
		}
	}
	return err
}

func (c *Config) NewClient(envConf *Config) (*APIClient, error) {
	if profileConf, err := c.LoadFromProfile(); err != nil {
		return nil, err
	} else {
		// 設定の優先度: tfファイル > 環境変数 > profile > プロバイダのデフォルト
		// ref: https://docs.usacloud.jp/terraform/provider/#api
		c.FillWith(envConf)
		c.FillWith(profileConf)
		c.FillWithDefault()
	}
	if err := c.validate(); err != nil {
		return nil, err
	}

	tfUserAgent := terraformUserAgent(c.TerraformVersion)
	providerUserAgent := fmt.Sprintf("%s/v%s", "terraform-provider-sakura", ver.Version)
	ua := fmt.Sprintf("%s %s", tfUserAgent, providerUserAgent)
	if add := os.Getenv(uaEnvVar); add != "" {
		ua += " " + add
		log.Printf("[DEBUG] Using modified User-Agent: %s", ua)
	}

	enableAPITrace := false
	enableHTTPTrace := false
	if c.TraceMode != "" {
		enableAPITrace = true
		enableHTTPTrace = true
		mode := strings.ToLower(c.TraceMode)
		switch mode {
		case traceAPI:
			enableHTTPTrace = false
		case traceHTTP:
			enableAPITrace = false
		}
	}
	callerOptions := &client.Options{
		AccessToken:          c.AccessToken,
		AccessTokenSecret:    c.AccessTokenSecret,
		AcceptLanguage:       c.AcceptLanguage,
		HttpRequestTimeout:   c.APIRequestTimeout,
		HttpRequestRateLimit: c.APIRequestRateLimit,
		RetryMax:             c.RetryMax,
		RetryWaitMax:         c.RetryWaitMax,
		RetryWaitMin:         c.RetryWaitMin,
		UserAgent:            ua,
		Trace:                enableHTTPTrace,
	}
	callerOptionsWithoutBigInt := &client.Options{
		AccessToken:          c.AccessToken,
		AccessTokenSecret:    c.AccessTokenSecret,
		AcceptLanguage:       c.AcceptLanguage,
		HttpRequestTimeout:   c.APIRequestTimeout,
		HttpRequestRateLimit: c.APIRequestRateLimit,
		RetryMax:             c.RetryMax,
		RetryWaitMax:         c.RetryWaitMax,
		RetryWaitMin:         c.RetryWaitMin,
		UserAgent:            ua,
		Trace:                enableHTTPTrace,
		RequestCustomizers: []saht.RequestCustomizer{
			func(req *http.Request) error {
				req.Header.Set("X-Sakura-Bigint-As-Int", "0")
				return nil
			}},
	}
	caller := api.NewCallerWithOptions(&api.CallerOptions{
		Options:     callerOptions,
		APIRootURL:  c.APIRootURL,
		DefaultZone: c.DefaultZone,
		TraceAPI:    enableAPITrace,
	})

	theClient := &saclient.Client{}
	if err := theClient.SetEnviron(c.createSaclientEnvConfig()); err != nil {
		return nil, fmt.Errorf("failed to create Sakura client via Envvars: %s", err.Error())
	}

	zones := c.Zones
	if len(zones) == 0 {
		zones = iaas.SakuraCloudZones
	}

	kmsClient, err := kms.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	smClient, err := sm.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	simplemqClient, err := simplemq.NewQueueClient(theClient)
	if err != nil {
		return nil, err
	}
	eventbusClient, err := eventbus.NewClient(client.WithOptions(callerOptions))
	if err != nil {
		return nil, err
	}
	nosqlClient, err := nosql.NewClient(client.WithOptions(callerOptions))
	if err != nil {
		return nil, err
	}
	dedicatedStorageClient, err := dedicatedstorage.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	apigwClient, err := apigw.NewClient(client.WithOptions(callerOptions))
	if err != nil {
		return nil, err
	}
	secconClient, err := seccon.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	iamClient, err := iam.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	simpleNotificationClient, err := simple_notification.NewClient(theClient)
	if err != nil {
		return nil, err
	}

	return &APIClient{
		APICaller:                        caller,
		defaultZone:                      c.Zone,
		zones:                            zones,
		deletionWaiterTimeout:            deletionWaiterTimeout,
		deletionWaiterPollingInterval:    deletionWaiterPollingInterval,
		databaseWaitAfterCreateDuration:  databaseWaitAfterCreateDuration,
		vpcRouterWaitAfterCreateDuration: vpcRouterWaitAfterCreateDuration,
		CallerOptions:                    callerOptions,
		KmsClient:                        kmsClient,
		SecretManagerClient:              smClient,
		SimpleMqClient:                   simplemqClient,
		EventBusClient:                   eventbusClient,
		AppRunClient:                     &apprun.Client{Options: callerOptions},
		ObjectStorageClient:              &objectstorage.Client{Options: callerOptionsWithoutBigInt},
		NosqlClient:                      nosqlClient,
		DedicatedStorageClient:           dedicatedStorageClient,
		ApigwClient:                      apigwClient,
		SecurityControlClient:            secconClient,
		IamClient:                        iamClient,
		SimpleNotificationClient:         simpleNotificationClient,
	}, nil
}

const tfUAEnvVar = "TF_APPEND_USER_AGENT"

func terraformUserAgent(version string) string {
	ua := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io)", version)

	if add := os.Getenv(tfUAEnvVar); add != "" {
		add = strings.TrimSpace(add)
		if len(add) > 0 {
			ua += " " + add
			log.Printf("[DEBUG] Using modified User-Agent: %s", ua)
		}
	}

	return ua
}

// saclient.Clientには内部storageのパラメータ群を直接設定する方法が現状ないので環境変数経由で設定する。
// profile -> 環境変数 -> TFファイルを全て適用した設定値を渡す
func (c *Config) createSaclientEnvConfig() []string {
	var envs []string

	if c.Profile != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_PROFILE=%s", c.Profile))
	}
	if c.AccessToken != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_ACCESS_TOKEN=%s", c.AccessToken))
	}
	if c.AccessTokenSecret != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_ACCESS_TOKEN_SECRET=%s", c.AccessTokenSecret))
	}
	if c.ServicePrincipalID != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_SERVICE_PRINCIPAL_ID=%s", c.ServicePrincipalID))
	}
	if c.ServicePrincipalKeyID != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_SERVICE_PRINCIPAL_KEY_ID=%s", c.ServicePrincipalKeyID))
	}
	if c.ServicePrivateKey != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_PRIVATE_KEY=%s", c.ServicePrivateKey))
	}
	if c.ServicePrivateKeyPath != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_PRIVATE_KEY_PATH=%s", c.ServicePrivateKeyPath))
	}
	if c.Zone != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_ZONE=%s", c.Zone))
	}
	if c.DefaultZone != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_DEFAULT_ZONE=%s", c.DefaultZone))
	}
	if len(c.Zones) > 0 {
		envs = append(envs, fmt.Sprintf("SAKURA_ZONES=%s", strings.Join(c.Zones, ",")))
	}
	if c.APIRootURL != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_API_ROOT_URL=%s", c.APIRootURL))
	}
	if c.RetryMax > 0 {
		envs = append(envs, fmt.Sprintf("SAKURA_RETRY_MAX=%d", c.RetryMax))
	}
	if c.RetryWaitMax > 0 {
		envs = append(envs, fmt.Sprintf("SAKURA_RETRY_WAIT_MAX=%d", c.RetryWaitMax))
	}
	if c.RetryWaitMin > 0 {
		envs = append(envs, fmt.Sprintf("SAKURA_RETRY_WAIT_MIN=%d", c.RetryWaitMin))
	}
	if c.APIRequestTimeout > 0 {
		envs = append(envs, fmt.Sprintf("SAKURA_API_REQUEST_TIMEOUT=%d", c.APIRequestTimeout))
	}
	if c.APIRequestRateLimit > 0 {
		envs = append(envs, fmt.Sprintf("SAKURA_RATE_LIMIT=%d", c.APIRequestRateLimit))
	}
	if c.TraceMode != "" {
		envs = append(envs, fmt.Sprintf("SAKURA_TRACE=%s", c.TraceMode))
	}
	if len(c.Endpoints) > 0 {
		for k, v := range c.Endpoints {
			envs = append(envs, fmt.Sprintf("%s%s=%s", EndpointsEnvVar, strings.ToUpper(k), v))
		}
	}

	return envs
}
