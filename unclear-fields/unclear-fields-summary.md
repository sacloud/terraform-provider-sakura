# Terraform Provider Sakura リソースドキュメント 不明瞭フィールド洗い出し

## 分析方法
- 対象: `docs/resources/` 配下の 138 リソース
- 各フィールドの Schema Description のみで「実際に入力すべき値が判断できない」ものを抽出
- Read-Only 属性は対象外

## 集計サマリー（おおよそ）

| カテゴリ | 件数（目安） |
|---|---|
| ID 形式不明 | 80+ |
| enum 値 / 選択肢不明 | 35+ |
| 数値の範囲 / step / 単位不明 | 55+ |
| IP / CIDR / ポート / URL / ホスト / パス形式不明 | 45+ |
| nested block の説明不足 | 25+ |
| 証明書 / 鍵 / シークレット / トークンの形式不明 | 20+ |
| JSON / YAML / CEL / Runbook スキーマ不明 | 10+ |
| Description が明らかに誤っている / 不自然 | 5+ |

## 優先度が高そうな項目（トップ 20）

| # | リソース | フィールド | 優先度の理由 |
|---|---|---|---|
| 1 | `sakura_iam_user` | `password_wo` | ~~Description が「Password for NoSQL appliance」と明らかに誤り~~ #292 |
| 2 | `sakura_simple_notification_destination` | `type` | Description が「ProcessConfiguration ID」と明らかに誤り |
| 3 | `sakura_simple_notification_group` | `destinations` | Description が「ProcessConfiguration ID」と明らかに誤り |
| 4 | `sakura_simple_notification_routing` | `match_labels` | Description が「The type of ...」と明らかに誤り |
| 5 | `sakura_server` | `private_host_name` | フィールド名は `name` だが Description が「id of the PrivateHost」と記載 |
| 6 | `sakura_archive` | `source_archive_zone` | フィールド名は `zone` だが Description が「share key」と矛盾 |
| 7 | `sakura_iam_policy` | `bindings.role.id` | Description が「The ID of the IAM Policy」と不自然（role の ID のはず） |
| 8 | `sakura_iam_organization_id_policy` | `bindings.role.id` | 同上 |
| 9 | `sakura_apprun_dedicated_version` | `registry_password_action` | enum 値が全く不明。入力できない |
| 10 | `sakura_apprun_dedicated_version` | `image` | container image の形式例がない（`nginx:latest` 等） |
| 11 | `sakura_apprun_dedicated_auto_scaling_group` | `worker_service_class_path` | path の形式・具体例がない |
| 12 | `sakura_apprun_dedicated_lb` | `service_class_path` | path の形式・具体例がない |
| 13 | `sakura_workflows` | `latest_revision.runbook` | YAML/JSON のスキーマ・必須構造が不明 |
| 14 | `sakura_security_control_automated_action` | `execution_condition` | CEL で使える変数・関数が不明 |
| 15 | `sakura_eventbus_process_configuration` | `parameters` | destination ごとの JSON スキーマが不明 |
| 16 | `sakura_monitoring_suite_alert_log_measure_rule` | `rule.query.matchers` | JSON スキーマが外部リンクのみで Description に具体例がない |
| 17 | `sakura_auto_scale` | `config` | sacloud/autoscaler 設定ファイルの形式/スキーマ/必須項目が不明 |
| 18 | `sakura_kms` | `plain_key` / `plain_key_wo` | キーの形式・長さ・エンコーディング（Base64/生バイナリ）が不明 |
| 19 | `sakura_apigw_service` | `oidc.id` | OIDC エンティティ ID の形式が不明 |
| 20 | `sakura_apigw_route` / `sakura_apigw_user` | `ip_restriction.*` | 許容される値（`allowIps`/`denyIps` 等）と CIDR の可否が不明 |

## カテゴリ別代表例

### 1. ID 形式不明（最も多い）
- `icon_id` はほとんど全リソースで「The icon id to attach to ...」とだけ記載（UUID/数値/その他が不明）
- `vswitch_id`, `upstream`, `switch.code`, `internet_id` など、どのリソースの ID か・形式が不明
- `project_id`, `subscription_id`, `plan_id`, `service_principal_id` など IAM/Workflows/API Gateway 系の ID
- `site_id`（Object Storage / WebAccel）はサイトコードか UUID か不明

### 2. enum 値 / 選択肢不明
- `sakura_apprun_dedicated_version.registry_password_action`
- `sakura_server.cpu_model`, `sakura_server.gpu_model`
- `sakura_vpn_router.version`
- `sakura_addon_*` の `location`（`japaneast` 以外に何があるか）
- `sakura_addon_*` の `sku`, `pricing_level`, `performance`, `redundancy`, `partition_count`, `replica_count`, `unit_count`
- `sakura_apigw_route` / `sakura_apigw_service` / `sakura_object_storage_*` の各種 enum 系

### 3. 数値の範囲 / step / 単位不明
- ノード/スケール系: `min_nodes`, `max_nodes`, `min_scale`, `max_scale`, `fixed_scale`
- サーバスペック: `core`, `memory`, `gpu`
- ネットマスク/VRID: `netmask`, `vrid`, `virtual_router_id`
- ポート番号: 各 `port` フィールド
- 時間/期間: `timeout`, `hour`, `retry_interval`, `retry_max`, `resend_interval_minutes`, `retention_period_days`

### 4. IP / CIDR / ポート / URL / ホスト / パス形式不明
- CIDR: `allowed_networks`, `source_ranges`, `source_network`, `ip_restriction.ips`
- IP: `name_servers`, `default_gateway`, `gateway`, `vip`, `network_address`, `ip_pool`
- URL/ホスト: `origin`, `endpoint`, `hostname`, `hosts`, `domain`
- パス: `worker_service_class_path`, `service_class_path`, `path`

### 5. nested block の説明不足
- `interfaces`（AppRun Dedicated）
- `network_interface`（database, read_replica, nfs, dsr_lb, vpn_router 等）
- `remark` / `settings`（nosql, nosql_additional_nodes）
- `conditions`（iam_auth, eventbus_trigger）
- `bindings`（iam_policy, iam_organization_id_policy）
- `action`（security_control_automated_action）
- `cors_config` / `object_storage_config` / `oidc`（apigw_service）

### 6. 証明書 / 鍵 / シークレット / トークンの形式不明
- `certificate_pem`, `private_key_pem`, `intermediate_certificate_pem`
- `idp_certificate`
- `public_key`, `ssh_keys`
- `secret_key`, `plain_key`, `access_key`, `secret_key`
- `onetime_url_secrets_wo`
- `authentication.jwt.key`

### 7. JSON / YAML / CEL / Runbook スキーマ不明
- `sakura_workflows.latest_revision.runbook`
- `sakura_security_control_automated_action.execution_condition`
- `sakura_eventbus_process_configuration.parameters`
- `sakura_monitoring_suite_alert_log_measure_rule.rule.query.matchers`
- `sakura_security_control_automated_action.action.parameters.args`
- `sakura_auto_scale.config`

### 8. Description が明らかに誤っている / 不自然
- `sakura_iam_user.password_wo`
- `sakura_simple_notification_destination.type`
- `sakura_simple_notification_group.destinations`
- `sakura_simple_notification_routing.match_labels`
- `sakura_server.private_host_name`
- `sakura_archive.source_archive_zone`
- `sakura_iam_policy.bindings.role.id`
- `sakura_iam_organization_id_policy.bindings.role.id`

## 次のステップ案
1. Description 誤り・不自然なものを優先して修正
2. enum 値が不明なものは実装側の Validator/PlanModifier から選択肢を確認し、ドキュメントに追記
3. ID 形式については、data source との関係や例を追記
4. nested block には構成例と各フィールドの説明を追記
5. 複雑な設定ファイル（JSON/YAML/CEL/Runbook）には最小限の例を追記

---

## 詳細ファイル

- Group 1: [AppRun / コンテナ / ミドルウェア](unclear-fields-detail-group1.md)
- Group 2: [ネットワーク](unclear-fields-detail-group2.md)
- Group 3: [コンピュート / ストレージ](unclear-fields-detail-group3.md)
- Group 4: [IAM / セキュリティ / 監視 / Workflows](unclear-fields-detail-group4.md)
- Group 5: [WebAccel / Addon / Object Storage / API Gateway](unclear-fields-detail-group5.md)
