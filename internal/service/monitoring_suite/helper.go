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

func getRoutingVariants(ctx context.Context, client *v1.Client) (map[string][]string, error) {
	api := monitoringsuite.NewPublisherOp(client)
	publishers, err := api.List(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list Monitoring Suite publishers: %w", err)
	}

	res := make(map[string][]string)
	for _, p := range publishers {
		vs := make([]string, 0, len(p.Variants))
		for _, v := range p.Variants {
			vs = append(vs, v.Name)
		}
		res[p.Code] = vs
	}
	return res, nil
}

func validateRoutingVariant(ctx context.Context, client *v1.Client, publisherCode, variantName string) error {
	variants, err := getRoutingVariants(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get routing variants: %w", err)
	}

	if vars, ok := variants[publisherCode]; ok {
		if !slices.Contains(vars, variantName) {
			return fmt.Errorf("The variant '%s' is not valid for publisher code '%s'. Valid variants are: %v", variantName, publisherCode, vars)
		}
	} else {
		return fmt.Errorf("The publisher code '%s' is not valid. Valid publisher codes are: %v", publisherCode, slices.Sorted(maps.Keys(variants)))
	}

	return nil
}
