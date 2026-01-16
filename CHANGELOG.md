# Changelog

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
