# Changelog

## [v3.3.0](https://github.com/sacloud/terraform-provider-sakura/compare/v3.2.1...v3.3.0) - 2026-02-01
- Add auto_scale by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/101
- Add security_control resources by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/98
- go: bump github.com/sacloud/api-client-go from 0.3.4 to 0.3.5 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/106
- Update NoSQL resource parameters with PlanModifiers by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/107

## [v3.2.1](https://github.com/sacloud/terraform-provider-sakura/compare/v3.2.0...v3.2.1) - 2026-01-27
- go: bump github.com/sacloud/saclient-go from 0.2.5 to 0.2.7 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/99
- ci: bump Songmu/tagpr from 1.10.0 to 1.11.1 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/94
- go: bump github.com/sacloud/apprun-api-go from 0.5.0 to 0.6.0 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/72
- go: bump github.com/minio/minio-go/v7 from 7.0.97 to 7.0.98 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/86
- go: bump github.com/sacloud/iaas-service-go from 1.21.0 to 1.21.1 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/88
- ci: bump actions/checkout from 6.0.1 to 6.0.2 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/97
- ci: bump actions/setup-go from 6.1.0 to 6.2.0 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/95
- Update code by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/103

## [v3.2.0](https://github.com/sacloud/terraform-provider-sakura/compare/v3.1.2...v3.2.0) - 2026-01-23
- Add apigw by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/92
- Add dsr_lb resources by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/93
- go: bump github.com/sacloud/kms-api-go from 0.3.0 to 0.3.1 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/74

## [v3.1.2](https://github.com/sacloud/terraform-provider-sakura/compare/v3.1.1...v3.1.2) - 2026-01-16
- provider: propagate extra settings to `theCLient` by @shyouhei in https://github.com/sacloud/terraform-provider-sakura/pull/83
- Run DedicatedStorage acceptance tests only when SAKURA_ENABLE_DEDICATED_STORAGE is set by @yamamoto-febc in https://github.com/sacloud/terraform-provider-sakura/pull/84
- Fix acceptance tests and ELB schema by @yamamoto-febc in https://github.com/sacloud/terraform-provider-sakura/pull/87
- test: fix enhanced LB data source acc test plan expectation by @yamamoto-febc in https://github.com/sacloud/terraform-provider-sakura/pull/89
- secret_manager_secret: Support write-only value by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/91

## [v3.1.1](https://github.com/sacloud/terraform-provider-sakura/compare/v3.1.0...v3.1.1) - 2026-01-09
- Update CHANGES and add write-only password guide by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/80
- disk: don't set empty string because dedicated_storage_id is optional by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/82

## [v3.1.0](https://github.com/sacloud/terraform-provider-sakura/compare/v3.0.1...v3.1.0) - 2026-01-08
- Add auto_backup resource by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/77
- Support write-only password by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/75
- feat: sakura_dedicated_storage by @yamamoto-febc in https://github.com/sacloud/terraform-provider-sakura/pull/78

## [v3.0.1](https://github.com/sacloud/terraform-provider-sakura/compare/v3.0.0...v3.0.1) - 2026-01-06
- Update README and CHANGES by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/63
- simple_mq: Fix description parameter handling by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/64
- Use tagpr for release by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/65
- feat: confidential vm by @yamamoto-febc in https://github.com/sacloud/terraform-provider-sakura/pull/61
- Refactor Makefile targets and add generate-docs target by @yamamoto-febc in https://github.com/sacloud/terraform-provider-sakura/pull/67
- eventbus_process_configuration: accept unknown state for credentials to support 'variable' by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/76

## [v3.0.0](https://github.com/sacloud/terraform-provider-sakura/compare/v3.0.0-rc9...v3.0.0) - 2025-12-23
- go: bump github.com/sacloud/api-client-go from 0.3.3 to 0.3.4 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/60
- ci: bump actions/checkout from 6.0.0 to 6.0.1 by @dependabot[bot] in https://github.com/sacloud/terraform-provider-sakura/pull/47
- v3.0.0 by @repeatedly in https://github.com/sacloud/terraform-provider-sakura/pull/62
