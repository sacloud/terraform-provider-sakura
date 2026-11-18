// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"time"

	"github.com/sacloud/sacloud-sdk-go/api/addon"
	addonapi "github.com/sacloud/sacloud-sdk-go/api/addon/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/api/apigw"
	apigwapi "github.com/sacloud/sacloud-sdk-go/api/apigw/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/api/apprun"
	apprun_dedicated "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated"
	apprundedicatedapi "github.com/sacloud/sacloud-sdk-go/api/apprun-dedicated/apis/v1"
	apprunapi "github.com/sacloud/sacloud-sdk-go/api/apprun/apis/v1"
	dedicatedstorage "github.com/sacloud/sacloud-sdk-go/api/dedicated-storage"
	dedicatedstorageapi "github.com/sacloud/sacloud-sdk-go/api/dedicated-storage/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/api/eventbus"
	eventbus_api "github.com/sacloud/sacloud-sdk-go/api/eventbus/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/api/iaas"
	"github.com/sacloud/sacloud-sdk-go/api/iaas/helper/query"
	"github.com/sacloud/sacloud-sdk-go/api/iam"
	iamapi "github.com/sacloud/sacloud-sdk-go/api/iam/apis/v1"
	kms "github.com/sacloud/sacloud-sdk-go/api/kms"
	kmsapi "github.com/sacloud/sacloud-sdk-go/api/kms/apis/v1"
	monitoringsuite "github.com/sacloud/sacloud-sdk-go/api/monitoring-suite"
	monitoringsuiteapi "github.com/sacloud/sacloud-sdk-go/api/monitoring-suite/apis/v1"
	nosql "github.com/sacloud/sacloud-sdk-go/api/nosql"
	nosqlapi "github.com/sacloud/sacloud-sdk-go/api/nosql/apis/v1"
	objectstorage "github.com/sacloud/sacloud-sdk-go/api/object-storage"
	sm "github.com/sacloud/sacloud-sdk-go/api/secretmanager"
	smapi "github.com/sacloud/sacloud-sdk-go/api/secretmanager/apis/v1"
	seccon "github.com/sacloud/sacloud-sdk-go/api/security-control"
	secconapi "github.com/sacloud/sacloud-sdk-go/api/security-control/apis/v1"
	simple_notification "github.com/sacloud/sacloud-sdk-go/api/simple-notification"
	simple_notification_api "github.com/sacloud/sacloud-sdk-go/api/simple-notification/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/api/simplemq"
	"github.com/sacloud/sacloud-sdk-go/api/simplemq/apis/v1/queue"
	"github.com/sacloud/sacloud-sdk-go/api/webaccel"
	"github.com/sacloud/sacloud-sdk-go/api/workflows"
	workflowsapi "github.com/sacloud/sacloud-sdk-go/api/workflows/apis/v1"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
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

// APIClient for SakuraCloud API
type APIClient struct {
	iaas.APICaller
	defaultZone                      string // 各リソースでzone未指定の場合に利用するゾーン。iaas.APIDefaultZoneとは別物。
	zones                            []string
	deletionWaiterTimeout            time.Duration
	deletionWaiterPollingInterval    time.Duration
	databaseWaitAfterCreateDuration  time.Duration
	vpcRouterWaitAfterCreateDuration time.Duration
	SaClient                         *saclient.Client
	AppRunClient                     *apprunapi.Client
	AppRunDedicatedClient            *apprundedicatedapi.Client
	KmsClient                        *kmsapi.Client
	SecretManagerClient              *smapi.Client
	SimpleMqClient                   *queue.Client
	EventBusClient                   *eventbus_api.Client
	ObjectStorageFedClient           *objectstorage.FedClient
	NosqlClient                      *nosqlapi.Client
	DedicatedStorageClient           *dedicatedstorageapi.Client
	ApigwClient                      *apigwapi.Client
	SecurityControlClient            *secconapi.Client
	IamClient                        *iamapi.Client
	AddonClient                      *addonapi.Client
	WorkflowsClient                  *workflowsapi.Client
	SimpleNotificationClient         *simple_notification_api.Client
	MonitoringSuiteClient            *monitoringsuiteapi.Client
	WebaccelClient                   *webaccel.Client
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

func NewClient(theClient *saclient.Client) (*APIClient, error) {
	c, err := theClient.EndpointConfig()
	if err != nil {
		return nil, err
	}

	caller := iaas.NewClientFromSaclient(theClient)

	zones := c.Zones
	if len(zones) == 0 {
		zones = iaas.SakuraCloudZones
	}

	apprundedicatedClient, err := apprun_dedicated.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	apprunClient, err := apprun.NewClient(theClient)
	if err != nil {
		return nil, err
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
	eventbusClient, err := eventbus.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	nosqlClient, err := nosql.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	dedicatedStorageClient, err := dedicatedstorage.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	apigwClient, err := apigw.NewClient(theClient)
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
	addonClient, err := addon.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	workflowsClient, err := workflows.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	simpleNotificationClient, err := simple_notification.NewClient(theClient)
	if err != nil {
		return nil, err
	}
	fedClient, err := objectstorage.NewFedClient(theClient)
	if err != nil {
		return nil, err
	}
	monitoringSuiteClient, err := monitoringsuite.NewClient(theClient)
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
		SaClient:                         theClient,
		KmsClient:                        kmsClient,
		SecretManagerClient:              smClient,
		SimpleMqClient:                   simplemqClient,
		EventBusClient:                   eventbusClient,
		AppRunClient:                     apprunClient,
		AppRunDedicatedClient:            apprundedicatedClient,
		ObjectStorageFedClient:           fedClient,
		NosqlClient:                      nosqlClient,
		DedicatedStorageClient:           dedicatedStorageClient,
		ApigwClient:                      apigwClient,
		SecurityControlClient:            secconClient,
		IamClient:                        iamClient,
		AddonClient:                      addonClient,
		WorkflowsClient:                  workflowsClient,
		SimpleNotificationClient:         simpleNotificationClient,
		MonitoringSuiteClient:            monitoringSuiteClient,
		WebaccelClient:                   &webaccel.Client{Saclient: theClient},
	}, nil
}
