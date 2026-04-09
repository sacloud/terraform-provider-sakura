# Copyright 2016-2026 The terraform-provider-sakura Authors
# SPDX-License-Identifier: Apache-2.0

terraform {
  required_providers {
    sakura = {
      source  = "sacloud/sakura"
      version = ">= 3.5.0" # needs IAM, ">= 3.5.0" is must.
    }
  }
}

provider "sakura" {
  # all params are set via ~/.usacloud
  profile = terraform.workspace
}
