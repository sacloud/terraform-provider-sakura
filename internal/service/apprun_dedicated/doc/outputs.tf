# Copyright 2016-2026 The terraform-provider-sakura Authors
# SPDX-License-Identifier: Apache-2.0

output "service_principal_id" {
  value       = sakura_iam_service_principal.execution.id
  description = "ID of the created or updated service principal, which is eligible for invoking apprun. This should be set to the environment variable SAKURA_APPRUN_DEDICATED_SERVICE_PRINCIPAL_ID for the tests to work."
}