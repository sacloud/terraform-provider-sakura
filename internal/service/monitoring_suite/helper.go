// Copyright 2016-2026 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package monitoring_suite

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	monitoringsuite "github.com/sacloud/monitoring-suite-api-go"
	v1 "github.com/sacloud/monitoring-suite-api-go/apis/v1"
)

func expandOptionalString(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueString()
	return &v
}

func parseUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid UUID: %w", err)
	}
	return id, nil
}

type routingVariant struct {
	Name string
	Type string
}

func getRoutingVariants(ctx context.Context, client *v1.Client) (map[string][]routingVariant, error) {
	api := monitoringsuite.NewPublisherOp(client)
	publishers, err := api.List(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list Monitoring Suite publishers: %w", err)
	}

	res := make(map[string][]routingVariant)
	for _, p := range publishers {
		vs := make([]routingVariant, 0, len(p.Variants))
		for _, v := range p.Variants {
			vs = append(vs, routingVariant{
				Name: v.Name,
				Type: string(v.Storage),
			})
		}
		res[p.Code] = vs
	}
	return res, nil
}

var storageTypeNames = map[string]string{
	"logs":    "sakura_monitoring_suite_log_storage",
	"metrics": "sakura_monitoring_suite_metric_storage",
}

func validateRoutingVariant(ctx context.Context, client *v1.Client, storageType, publisherCode, variantName string) error {
	routingVars, err := getRoutingVariants(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get routing variants: %w", err)
	}

	if variants, ok := routingVars[publisherCode]; ok {
		valid := false
		for _, v := range variants {
			if v.Name == variantName && v.Type == storageType {
				valid = true
				break
			}
		}
		if !valid {
			availableVars := make([]string, 0)
			for _, v := range variants {
				if v.Type == storageType {
					availableVars = append(availableVars, v.Name)
				}
			}
			if len(availableVars) == 0 {
				return fmt.Errorf("the publisher_code '%s' does not have any variants for '%s' resource type.", publisherCode, storageTypeNames[storageType])
			} else {
				return fmt.Errorf("the variant '%s' is not valid for publisher code '%s' with '%s' resource type. Valid variants are: %v", variantName, publisherCode, storageTypeNames[storageType], availableVars)
			}
		}
	} else {
		return fmt.Errorf("the publisher_code '%s' is not valid. Valid publisher codes are: %v", publisherCode, slices.Sorted(maps.Keys(routingVars)))
	}

	return nil
}
