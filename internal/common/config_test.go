// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/sacloud/api-client-go/profile"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
	"github.com/sacloud/terraform-provider-sakura/internal/defaults"
	"github.com/stretchr/testify/require"
)

func initTestProfileDir() func() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	profileDir := filepath.Join(wd, ".usacloud")
	os.Setenv("SAKURACLOUD_PROFILE_DIR", profileDir) //nolint
	if _, err := os.Stat(profileDir); err == nil {
		os.RemoveAll(profileDir) //nolint
	}

	return func() {
		os.RemoveAll(profileDir) //nolint
	}
}

func TestConfig_NewClient_loadFromProfile(t *testing.T) {
	defer initTestProfileDir()()

	defaultProfile := &saclient.Profile{
		Name: "default",
		Attributes: map[string]any{
			"AccessToken":          "token",
			"AccessTokenSecret":    "secret",
			"Zone":                 "dummy1",
			"Zones":                []string{"dummy1", "dummy2"},
			"UserAgent":            "dummy-ua",
			"AcceptLanguage":       "ja-JP",
			"RetryMax":             1,
			"RetryWaitMin":         2,
			"RetryWaitMax":         3,
			"StatePollingTimeout":  4,
			"StatePollingInterval": 5,
			"HTTPRequestTimeout":   6,
			"HTTPRequestRateLimit": 7,
			"APIRootURL":           "dummy",
			"TraceMode":            "dummy",
			"FakeMode":             true,
			"FakeStorePath":        "dummy",
		},
	}
	testProfile := &saclient.Profile{
		Name: "test",
		Attributes: map[string]any{
			"AccessToken":          "testtoken",
			"AccessTokenSecret":    "testsecret",
			"Zone":                 "test",
			"Zones":                []string{"test1", "test2"},
			"UserAgent":            "test-ua",
			"AcceptLanguage":       "ja-JP",
			"RetryMax":             7,
			"RetryWaitMin":         6,
			"RetryWaitMax":         5,
			"StatePollingTimeout":  4,
			"StatePollingInterval": 3,
			"HTTPRequestTimeout":   2,
			"HTTPRequestRateLimit": 1,
			"APIRootURL":           "test",
			"TraceMode":            "test",
			"FakeMode":             false,
			"FakeStorePath":        "test",
		},
	}

	// プロファイル指定なし & デフォルトプロファイルなし
	// プロファイル指定なし & デフォルトプロファイルあり
	// プロファイル指定あり & 指定プロファイルが存在しない
	// プロファイル指定あり 通常

	cases := []struct {
		scenario       string
		in             *common.Config
		profiles       map[string]*saclient.Profile
		expect         *common.Config
		currentProfile string
		err            error
	}{
		{
			scenario: "If profileName is not specified and profile is not exists, use default values",
			in: &common.Config{
				Profile:             "",
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				RetryMax:            defaults.RetryMax,
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
			profiles: map[string]*saclient.Profile{},
			expect: &common.Config{
				Profile:             "",
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				RetryMax:            defaults.RetryMax,
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
		},
		{
			scenario: "If no profile is specified and a current profile exists, it is loaded from the current profile",
			in: &common.Config{
				Profile: "",
			},
			profiles: map[string]*saclient.Profile{
				"default": defaultProfile,
				"test":    testProfile,
			},
			currentProfile: "test",
			expect: &common.Config{
				Profile:             "test",
				AccessToken:         testProfile.Attributes["AccessToken"].(string),
				AccessTokenSecret:   testProfile.Attributes["AccessTokenSecret"].(string),
				Zone:                testProfile.Attributes["Zone"].(string),
				Zones:               testProfile.Attributes["Zones"].([]string),
				TraceMode:           testProfile.Attributes["TraceMode"].(string),
				AcceptLanguage:      testProfile.Attributes["AcceptLanguage"].(string),
				APIRootURL:          testProfile.Attributes["APIRootURL"].(string),
				RetryMax:            testProfile.Attributes["RetryMax"].(int),
				RetryWaitMin:        testProfile.Attributes["RetryWaitMin"].(int),
				RetryWaitMax:        testProfile.Attributes["RetryWaitMax"].(int),
				APIRequestTimeout:   testProfile.Attributes["HTTPRequestTimeout"].(int),
				APIRequestRateLimit: testProfile.Attributes["HTTPRequestRateLimit"].(int),
			},
		},
		{
			scenario: "Values in the config are not overridden by the profile",
			in: &common.Config{
				Profile:           "",
				AccessToken:       "token",
				AccessTokenSecret: "secret",
				Zone:              "is1c",
			},
			profiles: map[string]*saclient.Profile{
				"default": defaultProfile,
				"test":    testProfile,
			},
			currentProfile: "test",
			expect: &common.Config{
				Profile:             "test",
				AccessToken:         "token",
				AccessTokenSecret:   "secret",
				Zone:                "is1c",
				Zones:               testProfile.Attributes["Zones"].([]string),
				TraceMode:           testProfile.Attributes["TraceMode"].(string),
				AcceptLanguage:      testProfile.Attributes["AcceptLanguage"].(string),
				APIRootURL:          testProfile.Attributes["APIRootURL"].(string),
				RetryMax:            testProfile.Attributes["RetryMax"].(int),
				RetryWaitMin:        testProfile.Attributes["RetryWaitMin"].(int),
				RetryWaitMax:        testProfile.Attributes["RetryWaitMax"].(int),
				APIRequestTimeout:   testProfile.Attributes["HTTPRequestTimeout"].(int),
				APIRequestRateLimit: testProfile.Attributes["HTTPRequestRateLimit"].(int),
			},
		},
		{
			scenario: "ProfileName is not specified and Profile is exists",
			in: &common.Config{
				Profile:             "",
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				RetryMax:            defaults.RetryMax,
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
			profiles: map[string]*saclient.Profile{
				"default": defaultProfile,
			},
			expect: &common.Config{
				Profile:             "default",
				AccessToken:         defaultProfile.Attributes["AccessToken"].(string),
				AccessTokenSecret:   defaultProfile.Attributes["AccessTokenSecret"].(string),
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				TraceMode:           defaultProfile.Attributes["TraceMode"].(string),
				AcceptLanguage:      defaultProfile.Attributes["AcceptLanguage"].(string),
				APIRootURL:          defaultProfile.Attributes["APIRootURL"].(string),
				RetryMax:            defaults.RetryMax,
				RetryWaitMin:        defaultProfile.Attributes["RetryWaitMin"].(int),
				RetryWaitMax:        defaultProfile.Attributes["RetryWaitMax"].(int),
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
		},
		{
			scenario: "Empty Config and Profile is exists",
			in: &common.Config{
				Profile:             "",
				Zone:                "",
				Zones:               nil,
				RetryMax:            0,
				APIRequestTimeout:   0,
				APIRequestRateLimit: 0,
			},
			profiles: map[string]*saclient.Profile{
				"default": defaultProfile,
			},
			expect: &common.Config{
				Profile:             "default",
				AccessToken:         defaultProfile.Attributes["AccessToken"].(string),
				AccessTokenSecret:   defaultProfile.Attributes["AccessTokenSecret"].(string),
				Zone:                defaultProfile.Attributes["Zone"].(string),
				Zones:               defaultProfile.Attributes["Zones"].([]string),
				TraceMode:           defaultProfile.Attributes["TraceMode"].(string),
				AcceptLanguage:      defaultProfile.Attributes["AcceptLanguage"].(string),
				APIRootURL:          defaultProfile.Attributes["APIRootURL"].(string),
				RetryMax:            defaultProfile.Attributes["RetryMax"].(int),
				RetryWaitMin:        defaultProfile.Attributes["RetryWaitMin"].(int),
				RetryWaitMax:        defaultProfile.Attributes["RetryWaitMax"].(int),
				APIRequestTimeout:   defaultProfile.Attributes["HTTPRequestTimeout"].(int),
				APIRequestRateLimit: defaultProfile.Attributes["HTTPRequestRateLimit"].(int),
			},
		},
		{
			scenario: "ProfileName is not specified with some values and Profile is exists",
			in: &common.Config{
				Profile:             "",
				AccessToken:         "from config",
				AccessTokenSecret:   "from config",
				Zone:                "from config",
				Zones:               []string{"zone1", "zone2"},
				TraceMode:           "from config",
				AcceptLanguage:      "from config",
				APIRootURL:          "from config",
				RetryMax:            8080,
				RetryWaitMin:        8080,
				RetryWaitMax:        8080,
				APIRequestTimeout:   8080,
				APIRequestRateLimit: 8080,
			},
			profiles: map[string]*saclient.Profile{
				"default": defaultProfile,
			},
			expect: &common.Config{
				Profile:             "default",
				AccessToken:         "from config",
				AccessTokenSecret:   "from config",
				Zone:                "from config",
				Zones:               []string{"zone1", "zone2"},
				TraceMode:           "from config",
				AcceptLanguage:      "from config",
				APIRootURL:          "from config",
				RetryMax:            8080,
				RetryWaitMin:        8080,
				RetryWaitMax:        8080,
				APIRequestTimeout:   8080,
				APIRequestRateLimit: 8080,
			},
		},
		{
			scenario: "Profile name specified but not exists",
			in: &common.Config{
				Profile: "test",
			},
			profiles: map[string]*saclient.Profile{
				"default": defaultProfile,
			},
			expect: &common.Config{
				Profile: "test",
			},
			err: errors.New(`failed to load profile[test]: API Error - failed to open test/config.json: openat test/config.json: no such file or directory`),
		},
		{
			scenario: "Profile name specified with normal profile",
			in: &common.Config{
				Profile:             "test",
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				RetryMax:            defaults.RetryMax,
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
			profiles: map[string]*saclient.Profile{
				"default": defaultProfile,
				"test":    testProfile,
			},
			expect: &common.Config{
				Profile:             "test",
				AccessToken:         testProfile.Attributes["AccessToken"].(string),
				AccessTokenSecret:   testProfile.Attributes["AccessTokenSecret"].(string),
				Zone:                defaults.Zone,
				Zones:               iaas.SakuraCloudZones,
				TraceMode:           testProfile.Attributes["TraceMode"].(string),
				AcceptLanguage:      testProfile.Attributes["AcceptLanguage"].(string),
				APIRootURL:          testProfile.Attributes["APIRootURL"].(string),
				RetryMax:            defaults.RetryMax,
				RetryWaitMin:        testProfile.Attributes["RetryWaitMin"].(int),
				RetryWaitMax:        testProfile.Attributes["RetryWaitMax"].(int),
				APIRequestTimeout:   defaults.APIRequestTimeout,
				APIRequestRateLimit: defaults.APIRequestRateLimit,
			},
		},
		{
			scenario: "only Profile name specified with normal profile",
			in: &common.Config{
				Profile: "test",
			},
			profiles: map[string]*saclient.Profile{
				"default": defaultProfile,
				"test":    testProfile,
			},
			expect: &common.Config{
				Profile:             "test",
				AccessToken:         testProfile.Attributes["AccessToken"].(string),
				AccessTokenSecret:   testProfile.Attributes["AccessTokenSecret"].(string),
				Zone:                testProfile.Attributes["Zone"].(string),
				Zones:               testProfile.Attributes["Zones"].([]string),
				TraceMode:           testProfile.Attributes["TraceMode"].(string),
				AcceptLanguage:      testProfile.Attributes["AcceptLanguage"].(string),
				APIRootURL:          testProfile.Attributes["APIRootURL"].(string),
				RetryMax:            testProfile.Attributes["RetryMax"].(int),
				RetryWaitMin:        testProfile.Attributes["RetryWaitMin"].(int),
				RetryWaitMax:        testProfile.Attributes["RetryWaitMax"].(int),
				APIRequestTimeout:   testProfile.Attributes["HTTPRequestTimeout"].(int),
				APIRequestRateLimit: testProfile.Attributes["HTTPRequestRateLimit"].(int),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.scenario, func(t *testing.T) {
			initTestProfileDir()
			profileOp := saclient.NewProfileOp(os.Environ())

			for _, profileValue := range tt.profiles {
				if err := profileOp.Create(profileValue); err != nil {
					t.Fatal(err)
				}
			}

			if len(tt.profiles) > 0 {
				currentProfile := tt.currentProfile
				if tt.currentProfile == "" {
					currentProfile = profile.DefaultProfileName
				}
				if err := profileOp.SetCurrentName(currentProfile); err != nil {
					t.Fatal(err)
				}
			}

			cfg, err := tt.in.LoadFromProfile()
			if err != nil {
				if tt.err.Error() != err.Error() {
					t.Errorf("got unexpected error: expected: %s got: %s", tt.err, err)
				}
			} else {
				tt.in.FillWith(cfg)
				require.EqualValues(t, tt.expect, tt.in)
			}
		})
	}
}
