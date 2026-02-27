// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package addon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v1 "github.com/sacloud/addon-api-go/apis/v1"
	"github.com/sacloud/saclient-go"
	"github.com/sacloud/terraform-provider-sakura/internal/common"
)

func configureAddonClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) *v1.Client {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return nil
	}
	return apiclient.AddonClient
}

func configureAddonDataSourceClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *v1.Client {
	apiclient := common.GetApiClientFromProvider(req.ProviderData, &resp.Diagnostics)
	if apiclient == nil {
		return nil
	}
	return apiclient.AddonClient
}

func getAddon(ctx context.Context, name, id string, read func(context.Context, string) (*v1.GetResourceResponse, error), state *tfsdk.State, diags *diag.Diagnostics) *v1.GetResourceResponse {
	resp, err := read(ctx, id)
	if err != nil {
		if saclient.IsNotFoundError(err) {
			state.RemoveResource(ctx)
			return nil
		}
		diags.AddError("API Read Error", fmt.Sprintf("failed to read Addon %s[%s]: %s", name, id, err))
		return nil
	}
	return resp
}

func getAddonIDsFromDeployment(name string, resp *v1.PostDeploymentResponse, diags *diag.Diagnostics) (string, string, bool) {
	if resp == nil {
		diags.AddError("Create: API Error", fmt.Sprintf("failed to create Addon %s: empty response", name))
		return "", "", false
	}

	resourceGroupName, ok := resp.ResourceGroupName.Get()
	if !ok || resourceGroupName == "" {
		diags.AddError("Create: API Error", fmt.Sprintf("failed to create Addon %s: missing resource group name", name))
		return "", "", false
	}

	deploymentName, _ := resp.DeploymentName.Get()
	return resourceGroupName, deploymentName, true
}

func deploymentNameValue(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

// Japan East -> japaneast
// この変換が全てのロケーションで正しいかは不明だが、現状はこれで対応する
func loweredLocation(value string) string {
	words := []string{}
	for _, word := range strings.Fields(value) {
		words = append(words, strings.ToLower(word))
	}
	return strings.Join(words, "")
}

func getCDNLocationAndProfile(data map[string]any) (string, v1.FrontDoorProfile, error) {
	var profile v1.FrontDoorProfile
	location, ok := data["location"].(string)
	if !ok || location == "" {
		return "", profile, errors.New("missing location in addon response")
	}
	sku, ok := data["sku"].(map[string]any)
	if !ok {
		return "", profile, errors.New("missing sku in addon response")
	}
	skuName, ok := sku["name"].(string)
	if !ok || skuName == "" {
		return "", profile, errors.New("missing sku name in addon response")
	}
	parts := strings.SplitN(skuName, "_", 2)
	level, ok := pricingLevelMap[parts[0]]
	if !ok {
		return "", profile, fmt.Errorf("unknown pricing level: %s", parts[0])
	}
	profile.Level = level
	return location, profile, nil
}

func getFrontDoorEndpoint(data map[string]any) (v1.FrontDoorEndpoint, error) {
	var endpoint v1.FrontDoorEndpoint
	hostName, hostHeader, patterns, err := getCDNFamilyValues(data)
	if err != nil {
		return endpoint, err
	}
	endpoint.Route.Patterns = patterns
	endpoint.Route.OriginGroup.Origin.HostName = hostName
	endpoint.Route.OriginGroup.Origin.HostHeader = hostHeader
	return endpoint, nil
}

func getCDNFamilyValues(data map[string]any) (string, string, []string, error) {
	patterns := []string{}
	endpoints, ok := data["endpoints"].([]any)
	if !ok || len(endpoints) == 0 {
		return "", "", nil, errors.New("missing endpoints in addon response")
	}
	endpoint, ok := endpoints[0].(map[string]any)
	if !ok {
		return "", "", nil, errors.New("invalid endpoints in addon response")
	}
	if routes, ok := endpoint["routes"].([]any); ok && len(routes) > 0 {
		if route, ok := routes[0].(map[string]any); ok {
			if properties, ok := route["properties"].(map[string]any); ok {
				if patternsToMatch, ok := properties["patternsToMatch"].([]any); ok {
					for _, p := range patternsToMatch {
						if pattern, ok := p.(string); ok {
							patterns = append(patterns, pattern)
						}
					}
				}
			}
		}
	}

	originGroups, ok := data["originGroups"].([]any)
	if !ok || len(originGroups) == 0 {
		return "", "", nil, errors.New("missing origin groups in addon response")
	}
	originGroup, ok := originGroups[0].(map[string]any)
	if !ok {
		return "", "", nil, errors.New("invalid origin group in addon response")
	}
	origins, ok := originGroup["origins"].([]any)
	if !ok || len(origins) == 0 {
		return "", "", nil, errors.New("missing origins in addon response")
	}
	origin, ok := origins[0].(map[string]any)
	if !ok {
		return "", "", nil, errors.New("invalid origin in addon response")
	}
	properties, ok := origin["properties"].(map[string]any)
	if !ok {
		return "", "", nil, errors.New("invalid origin properties in addon response")
	}
	hostName, ok := properties["hostName"].(string)
	if !ok || hostName == "" {
		return "", "", nil, errors.New("missing origin host name in addon response")
	}
	hostHeader, _ := properties["originHostHeader"].(string)
	return hostName, hostHeader, patterns, nil
}

func decodeCDNFamilyResponse[T interface {
	SetLocation(val string)
	SetProfile(val v1.FrontDoorProfile)
	SetEndpoint(val v1.FrontDoorEndpoint)
}](resp *v1.GetResourceResponse, result T) error {
	if resp == nil || len(resp.Data) == 0 {
		return errors.New("got invalid response from Addon CDN API")
	}

	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return err
	}

	location, profile, err := getCDNLocationAndProfile(data)
	if err != nil {
		return err
	}
	endpoint, err := getFrontDoorEndpoint(data)
	if err != nil {
		return err
	}
	result.SetLocation(location)
	result.SetProfile(profile)
	result.SetEndpoint(endpoint)

	return nil
}

func waitDeployment(ctx context.Context, name string, read func(context.Context, string) (*v1.GetResourceResponse, error), id string) (*v1.GetResourceResponse, error) {
	errCount := 0
	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	for {
		select {
		case <-waitCtx.Done():
			return nil, fmt.Errorf("timeout exceeded for Addon %s[%s] ready check: %s", name, id, waitCtx.Err())
		default:
			res, err := read(ctx, id)
			if err != nil {
				// 作成直後は404が返ることがあるため、404の場合はエラーとせずにリトライする
				if saclient.IsNotFoundError(err) {
					time.Sleep(10 * time.Second)
					continue
				}

				errCount += 1
				if errCount > 5 {
					return nil, fmt.Errorf("exceeds 5 retry limit during Addon %s[%s] ready check: %w", name, id, err)
				}

				time.Sleep(15 * time.Second)
				continue
			}

			return res, nil
		}
	}
}

func waitCDNRouteDeployment(ctx context.Context, read func(context.Context, string) (*v1.GetResourceResponse, error), id string) error {
	errCount := 0
	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	for {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("timeout exceeded for Addon CDN[%s] ready check: %s", id, waitCtx.Err())
		default:
			res, err := read(ctx, id)
			if err != nil {
				errCount += 1
				if errCount > 5 {
					return fmt.Errorf("exceeds 5 retry limit during Addon CDN[%s] ready check: %w", id, err)
				}
				time.Sleep(10 * time.Second)
				continue
			}

			var data map[string]any
			if err := json.Unmarshal(res.Data, &data); err != nil {
				return fmt.Errorf("failed to decode Addon CDN[%s] response: %w", id, err)
			}
			if len(data["endpoints"].([]any)) > 0 && len(data["endpoints"].([]any)[0].(map[string]any)["routes"].([]any)) > 0 {
				return nil
			}

			time.Sleep(15 * time.Second)
		}
	}
}
