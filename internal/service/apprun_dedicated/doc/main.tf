# Copyright 2016-2026 The terraform-provider-sakura Authors
# SPDX-License-Identifier: Apache-2.0

resource "random_id" "main" {
  byte_length = 8
  prefix      = "tfacc-m"
}

resource "random_id" "execution" {
  byte_length = 8
  prefix      = "tfacc-e"
}

resource "random_id" "configuration" {
  byte_length = 8
  prefix      = "tfacc-c"
}

resource "sakura_iam_project" "main" {
  code        = random_id.main.hex
  name        = random_id.main.hex
  description = "Terraform integration test project for AppRun"
}

resource "sakura_iam_service_principal" "execution" {
  name        = random_id.execution.hex
  project_id  = sakura_iam_project.main.id
  description = "This is used when AppRun scales up/down"
}

resource "sakura_iam_service_principal" "configuration" {
  name        = random_id.configuration.hex
  project_id  = sakura_iam_project.main.id
  description = "This is used when terraform sets AppRun up."
}

resource "sakura_iam_policy" "main" {
  target    = "project"
  target_id = sakura_iam_project.main.id

  bindings = [
    {
      role = {
        type = "preset"
        id   = "apprun-admin"
      }
      principals = [
        {
          type = "service-principal"
          id   = sakura_iam_service_principal.configuration.id
        }
      ]
    },
    {
      role = {
        type = "preset"
        id   = "resource-creator"
      }
      principals = [
        {
          type = "service-principal"
          id   = sakura_iam_service_principal.execution.id
        },
        {
          type = "service-principal"
          id   = sakura_iam_service_principal.configuration.id
        }
      ]
    },
  ]
}