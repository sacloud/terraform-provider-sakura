// Copyright 2016-2025 terraform-provider-sakuracloud authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/terraform-provider-sakuracloud/internal/common"
)

func CheckSakuraDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource is not exists: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("id is not set: %s", n)
		}
		return nil
	}
}

func CheckSakuraSwitchDestroy(s *terraform.State) error {
	swOp := iaas.NewSwitchOp(AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_switch" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		zone := rs.Primary.Attributes["zone"]
		_, err := swOp.Read(context.Background(), zone, common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("resource Switch[%s] still exists", rs.Primary.ID)
		}
	}

	return nil
}

func CheckSakuraIconDestroy(s *terraform.State) error {
	iconOp := iaas.NewIconOp(AccClientGetter())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "sakura_icon" {
			continue
		}
		if rs.Primary.ID == "" {
			continue
		}

		_, err := iconOp.Read(context.Background(), common.SakuraCloudID(rs.Primary.ID))
		if err == nil {
			return fmt.Errorf("still exists Icon: %s", rs.Primary.ID)
		}
	}

	return nil
}
