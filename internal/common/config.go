// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"

	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/api-client-go/profile"
	"github.com/sacloud/apprun-api-go"
	"github.com/sacloud/eventbus-api-go"
	eventbus_api "github.com/sacloud/eventbus-api-go/apis/v1"
	saht "github.com/sacloud/go-http"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/api"
	"github.com/sacloud/iaas-api-go/helper/query"
	objectstorage "github.com/sacloud/object-storage-api-go"
	"github.com/sacloud/simplemq-api-go"
	"github.com/sacloud/simplemq-api-go/apis/v1/queue"

	"github.com/sacloud/cloudhsm-api-go"
	cloudhsmapi "github.com/sacloud/cloudhsm-api-go/apis/v1"
	kms "github.com/sacloud/kms-api-go"
	kmsapi "github.com/sacloud/kms-api-go/apis/v1"
	sm "github.com/sacloud/secretmanager-api-go"
	smapi "github.com/sacloud/secretmanager-api-go/apis/v1"
	ver "github.com/sacloud/terraform-provider-sakura/version"
)

const (
	traceHTTP = "http"
	traceAPI  = "api"
)

const uaEnvVar = "SAKURACLOUD_APPEND_USER_AGENT"

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
	Profile             string
	AccessToken         string
	AccessTokenSecret   string
	Zone                string
	Zones               []string
	DefaultZone         string
	TraceMode           string
	AcceptLanguage      string
	APIRootURL          string
	RetryMax            int
	RetryWaitMin        int
	RetryWaitMax        int
	APIRequestTimeout   int
	APIRequestRateLimit int
	TerraformVersion    string
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
	CloudHSMClient                   *cloudhsmapi.Client
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

func (c *Config) loadFromProfile() error {
	if c.Profile == "" {
		c.Profile = profile.DefaultProfileName
	}
	if c.Profile != profile.DefaultProfileName {
		log.Printf("[DEBUG] using profile %q", c.Profile)
	}

	pcv := &profile.ConfigValue{}
	if err := profile.Load(c.Profile, pcv); err != nil {
		return fmt.Errorf("loading profile %q is failed: %s", c.Profile, err)
	}

	if c.AccessToken == "" {
		c.AccessToken = pcv.AccessToken
	}
	if c.AccessTokenSecret == "" {
		c.AccessTokenSecret = pcv.AccessTokenSecret
	}
	if (c.Zone == "" || c.Zone == Zone) && pcv.Zone != "" {
		c.Zone = pcv.Zone
	}

	defaultZones := iaas.SakuraCloudZones
	sort.Strings(c.Zones)
	sort.Strings(defaultZones)
	if (len(c.Zones) == 0 || reflect.DeepEqual(defaultZones, c.Zones)) && len(pcv.Zones) > 0 {
		c.Zones = pcv.Zones
	}
	if c.TraceMode == "" {
		c.TraceMode = pcv.TraceMode
	}
	if c.AcceptLanguage == "" {
		c.AcceptLanguage = pcv.AcceptLanguage
	}
	if c.APIRootURL == "" {
		c.APIRootURL = pcv.APIRootURL
	}
	if (c.RetryMax == 0 || c.RetryMax == RetryMax) && pcv.RetryMax > 0 {
		c.RetryMax = pcv.RetryMax
	}
	if c.RetryWaitMax == 0 {
		c.RetryWaitMax = pcv.RetryWaitMax
	}
	if c.RetryWaitMin == 0 {
		c.RetryWaitMin = pcv.RetryWaitMin
	}
	if (c.APIRequestTimeout == 0 || c.APIRequestTimeout == APIRequestTimeout) && pcv.HTTPRequestTimeout > 0 {
		c.APIRequestTimeout = pcv.HTTPRequestTimeout
	}
	if (c.APIRequestRateLimit == 0 || c.APIRequestRateLimit == APIRequestRateLimit) && pcv.HTTPRequestRateLimit > 0 {
		c.APIRequestRateLimit = pcv.HTTPRequestRateLimit
	}
	return nil
}

func (c *Config) validate() error {
	var err error
	if c.AccessToken == "" {
		err = multierror.Append(err, errors.New("AccessToken is required"))
	}
	if c.AccessTokenSecret == "" {
		err = multierror.Append(err, errors.New("AccessTokenSecret is required"))
	}
	return err
}

// NewClient returns new API Client for SakuraCloud
func (c *Config) NewClient() (*APIClient, error) {
	if err := c.loadFromProfile(); err != nil {
		return nil, err
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
		HttpClient:           http.DefaultClient,
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
		HttpClient:           http.DefaultClient,
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

	zones := c.Zones
	if len(zones) == 0 {
		zones = iaas.SakuraCloudZones
	}

	kmsClient, err := kms.NewClient(client.WithOptions(callerOptions))
	if err != nil {
		return nil, err
	}
	smClient, err := sm.NewClient(client.WithOptions(callerOptions))
	if err != nil {
		return nil, err
	}
	simplemqClient, err := simplemq.NewQueueClient(client.WithOptions(callerOptions))
	if err != nil {
		return nil, err
	}
	eventbusClient, err := eventbus.NewClient(client.WithOptions(callerOptions))
	if err != nil {
		return nil, err
	}
	cloudhsmClient, err := cloudhsm.NewClient(client.WithOptions(callerOptions))
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
		CloudHSMClient:                   cloudhsmClient,
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
