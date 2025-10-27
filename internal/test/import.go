// Copyright 2016-2025 The terraform-provider-sakura Authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func CompareState(s *terraform.InstanceState, key, value string) error {
	actual := s.Attributes[key]
	if actual != value {
		return fmt.Errorf("expected state[%s] is %q, but %q received",
			key, value, actual)
	}
	return nil
}

func CompareStateMulti(s *terraform.InstanceState, expects map[string]string) error {
	for k, v := range expects {
		err := CompareState(s, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func StateNotEmpty(s *terraform.InstanceState, key string) error {
	if v, ok := s.Attributes[key]; !ok || v == "" {
		return fmt.Errorf("state[%s] is expected not empty", key)
	}
	return nil
}

func StateNotEmptyMulti(s *terraform.InstanceState, keys ...string) error {
	for _, key := range keys {
		if err := StateNotEmpty(s, key); err != nil {
			return err
		}
	}
	return nil
}
