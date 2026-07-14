# IAM / セキュリティ / 監視 / Workflows リソースの不明瞭フィールド詳細

## sakura_iam_auth

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `conditions` | Attributes | （Description なし） | どのような条件を設定できるか全く記載がない |
| `conditions.datetime_restriction.after` | String | The start time for datetime restriction. | 日時書式の明示がない（例はあるが Description にない） |
| `conditions.datetime_restriction.before` | String | The end time for datetime restriction. | 同上 |
| `conditions.ip_restriction.mode` | String | The mode of IP restriction. | 取りうる値の列挙がない（例は `allow_list` のみ） |
| `conditions.ip_restriction.source_network` | List of String | The source networks for IP restriction. | IP アドレスなのか CIDR なのか、形式が不明 |
| `password_policy.min_length` | Number | The minimum length of the password. | 最小値・最大値、step が不明 |

## sakura_iam_folder

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `parent_id` | String | The parent folder ID of IAM Folder. | ID の形式（UUID か resource_id か）が不明 |

## sakura_iam_group

問題なし

## sakura_iam_organization_id_policy

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `bindings.principals.id` | String | The ID of the principal | principal の種別と ID 形式が不明 |
| `bindings.principals.type` | String | The type of the principal | 取りうる値の列挙がない（例は `service-principal` のみ） |
| `bindings.role.id` | String | The ID of the IAM Organization ID Policy | 説明が不自然（role の ID のはず）。プリセット ID か何か、形式も不明 |
| `bindings.role.type` | String | The type of the role | 取りうる値の列挙がない（例は `preset` のみ） |

## sakura_iam_policy

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `target_id` | String | The ID of the target. Required for Folder or Project | 条件付き Required の具体例、ID 形式が不明 |
| `bindings.principals.id` | String | The ID of the principal | principal の種別と ID 形式が不明 |
| `bindings.principals.type` | String | The type of the principal | 取りうる値の列挙がない |
| `bindings.role.id` | String | The ID of the IAM Policy | 説明が不自然（role の ID のはず）。形式も不明 |
| `bindings.role.type` | String | The type of the role | 取りうる値の列挙がない |

## sakura_iam_project

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `code` | String | The code of the IAM Project | 形式や文字種、長さ制限が不明 |
| `parent_folder_id` | String | The parent folder ID of IAM Project. | ID 形式が不明 |

## sakura_iam_project_apikey

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `iam_roles` | List of String | The IAM roles assigned to the IAM Project API Key | 取りうるロール名の列挙や例がない |
| `project_id` | String | The project ID associated with the IAM Project API Key | ID 形式（resource_id / UUID 等）が不明 |
| `server_resource_id` | String | The server resource ID of IAM Project API Key. | 何を指すか、形式が不明 |
| `zone` | String | The zone of IAM Project API Key. | ゾーン名の例や列挙がない |

## sakura_iam_service_principal

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `project_id` | String | The project ID associated with the IAM Service Principal | ID 形式が不明 |

## sakura_iam_sso

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `idp_certificate` | String | The IdP certificate of the IAM SSO | 証明書の形式（PEM 全文かファイルパスか）が Description にない |

## sakura_iam_user

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `code` | String | The code of the IAM User | 形式や文字種、長さ制限が不明 |
| `password_wo` | String | Password for NoSQL appliance | 明らかに説明が間違っている（IAM User のパスワードのはず） |

## sakura_iam_user_provisioning

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `token_version` | Number | The version of secret_token. Increment this to regenerate the token. | 取りうる範囲や初期値の例がない |

## sakura_security_control_activation

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `service_principal_id` | String | The Service Principal ID associated with the Project | ID 形式が不明 |

## sakura_security_control_automated_action

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `action.parameters.service_principal_id` | String | The Service Principal ID associated with the Automated Action | ID 形式が不明 |
| `action.parameters.target_id` | String | The id of target resource for the Automated Action | どのリソースの ID か、形式が不明 |
| `action.parameters.args` | String | The json formatted arguments to be passed to the workflow | JSON のスキーマや必須キーが不明 |
| `action.parameters.revision` | Number | The revision number of workflow to be executed | 範囲や指定方法が不明 |
| `action.parameters.revision_alias` | String | The revision alias of workflow to be executed | 取りうる値や例が不明 |
| `execution_condition` | String | The CEL expression that defines the condition for Automated Action trigger | CEL で使える変数・関数が不明 |

## sakura_security_control_evaluation_rule

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `id` | String | The ID of the Evaluation Rule. ... can be found in the documentation. | 利用可能な ID 一覧へのリンクがない（例はある） |
| `parameters.service_principal_id` | String | The Service Principal ID associated with the Evaluation Rule | ID 形式が不明 |
| `parameters.targets` | List of String | The list of targets for the Evaluation Rule | zone 名か site 名か、形式が不明（例は zone コード） |

## sakura_monitoring_suite_alert_log_measure_rule

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `alert_project_id` | String | The resource ID of the Alert Project. | resource ID の形式が不明 |
| `log_storage_id` | String | The resource ID of the Log Storage. | 同上 |
| `metric_storage_id` | String | The resource ID of the Metric Storage. | 同上 |
| `rule.version` | String | The version of the rule. | 取りうる値の列挙や例がない |
| `rule.query.matchers` | String | The matchers of the query in JSON format. | JSON スキーマが外部リンクのみで、Description 内に具体例がない |

## sakura_monitoring_suite_alert_notification_routing

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `alert_project_id` | String | The resource ID of the Alert Project. | resource ID の形式が不明 |
| `notification_target_id` | String | The UUID based ID of the Alert Notification Target. | 形式は UUID とあるが、例では resource_id 形式を示唆しており矛盾がある |
| `resend_interval_minutes` | Number | The resend interval in minutes of the Alert Notification Routing. | 範囲や step が不明 |

## sakura_monitoring_suite_alert_notification_target

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `alert_project_id` | String | The resource ID of the Alert Project. | resource ID の形式が不明 |
| `service_type` | String | The service type of the Alert Notification Target. | 取りうる値の列挙がない（例のみ） |
| `url` | String | The URL of the Alert Notification Target. | どの service_type で必要か、URL の形式が不明 |

## sakura_monitoring_suite_alert_project

問題なし

## sakura_monitoring_suite_alert_rule

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `alert_project_id` | String | The resource ID of the Alert Project. | resource ID の形式が不明 |
| `metric_storage_id` | String | The resource ID of the Metric Storage. | 同上 |
| `query` | String | The query of the Alert Rule. | 取りうる値の列挙や例がない（例は `count_values` のみ） |
| `format` | String | The format of the Alert Rule. | 取りうる値の列挙がない |
| `template` | String | The template of the Alert Rule. | テンプレート書式や例がない |
| `threshold_critical` | String | The threshold of critical level of the Alert Rule. | 書式例（例はあるが Description にない） |
| `threshold_warning` | String | The threshold of warning level of the Alert Rule. | 同上 |

## sakura_monitoring_suite_dashboard

問題なし

## sakura_monitoring_suite_log_routing

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `publisher_code` | String | The publisher code of the target service. | 取りうる値の列挙が Description にない（例と外部リンクのみ） |
| `storage_id` | String | The resource ID of the Log Storage. | resource ID の形式が不明 |
| `variant` | String | The variant of the Log Routing. | 取りうる値の列挙がない（例のみ） |
| `resource_id` | String | The resource ID of the target service. | resource ID の形式が不明 |

## sakura_monitoring_suite_log_storage

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `classification` | String | The bucket classification of the Log Storage. | 取りうる値の列挙がない（例は `shared`/`dedicated` のみ） |
| `retention_period_days` | Number | The retention period days of the Log Storage. | 範囲やデフォルト値が不明（例コメントのみ） |

## sakura_monitoring_suite_log_storage_access_key

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `storage_id` | String | The Log Storage ID for the Access Key. | ID 形式（resource_id / UUID 等）が不明 |

## sakura_monitoring_suite_metric_routing

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `publisher_code` | String | The publisher code of the target service. | 取りうる値の列挙が Description にない |
| `storage_id` | String | The resource ID of the Metric Storage. | resource ID の形式が不明 |
| `variant` | String | The variant of the Metric Routing. | 取りうる値の列挙がない |
| `resource_id` | String | The resource ID of the target service. | resource ID の形式が不明 |

## sakura_monitoring_suite_metric_storage

問題なし

## sakura_monitoring_suite_metric_storage_access_key

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `storage_id` | String | The Metric Storage ID for the Access Key. | ID 形式が不明 |

## sakura_monitoring_suite_trace_storage

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `retention_period_days` | Number | The retention period days of the Trace Storage. | 範囲やデフォルト値が不明 |

## sakura_monitoring_suite_trace_storage_access_key

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `storage_id` | String | The Trace Storage ID for the Access Key. | ID 形式が不明 |

## sakura_simple_monitor

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `health_check.ftps` | String | The methods of invoking security for monitoring with FTPS. | 取りうる値の列挙がない |
| `health_check.oid` | String | The SNMP OID used when checking by SNMP | OID 書式の例がない |
| `health_check.expected_data` | String | The expected value used when checking by DNS | DNS レコードのどの値と比較するか、形式が不明 |
| `health_check.remaining_days` | Number | The number of remaining days until certificate expiration used when checking SSL certificates. | 範囲が不明 |

## sakura_simple_notification_destination

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `type` | String | The ProcessConfiguration ID of the SimpleNotification Destination. | 説明が明らかに誤り（`email` 等のタイプのはず）。取りうる値の列挙がない |
| `value` | String | The source of the SimpleNotification Destination. | `type` に対応する値の形式や例がない |

## sakura_simple_notification_group

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `destinations` | List of String | The ProcessConfiguration ID of the SimpleNotification group. | 説明が誤り（destination ID のはず）。ID 形式も不明 |

## sakura_simple_notification_routing

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `match_labels` | Attributes List | The type of the SimpleNotification Routing. | 説明が明らかに誤り |
| `source_id` | String | The ID of the service that generates notifications. Available IDs can be retrieved via the API | 利用可能な値の列挙や具体例がない |
| `target_group_id` | String | The value of the simple_notification_group id | ID 形式が不明 |

## sakura_workflows

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `subscription_id` | String | The subscription ID of the Workflows. | ID 形式が不明 |
| `concurrency_mode` | String | The concurrency mode of the Workflows. | 取りうる値の列挙や例がない |
| `service_principal_id` | String | The service principal id of the Workflows. | ID 形式が不明 |
| `latest_revision.runbook` | String | The runbook definition of the revision. | YAML/JSON のスキーマや必須構造が不明 |

## sakura_workflows_revision_alias

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `workflow_id` | String | The workflow ID of the Workflows RevisionAlias. | ID 形式が不明 |
| `revision_id` | String | The revision ID of the Workflows RevisionAlias. | ID 形式が不明 |
| `alias` | String | The alias name of the Workflows RevisionAlias. | 取りうる値や予約語の有無が不明 |

## sakura_workflows_subscription

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `plan_id` | String | The plan ID of the Workflows Subscription. | ID 形式が不明 |

## sakura_eventbus_process_configuration

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `destination` | String | The destination of the EventBus ProcessConfiguration. | 取りうる値の列挙がない（例のみ） |
| `parameters` | String | The parameter of the EventBus ProcessConfiguration. | destination ごとの JSON スキーマが不明 |
| `credentials_wo_version` | Number | Version number for credentials. Change this when changing credentials. | 範囲や初期値が不明 |
| `sakura_access_token_wo` | String | The SimpleNotification/AutoScale access token ... | どの destination で必須か不明 |
| `sakura_access_token_secret_wo` | String | The SimpleNotification/AutoScale access token secret ... | 同上 |
| `simplemq_api_key_wo` | String | The SimpleMQ API key for EventBus ProcessConfiguration. | どの destination で必須か不明 |

## sakura_eventbus_schedule

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `process_configuration_id` | String | The ProcessConfiguration ID of the EventBus Schedule. | ID 形式が不明 |
| `crontab` | String | Crontab of the EventBus Schedule. | 書式（5 項目か 6 項目か、拡張表現の有無）が不明 |
| `recurring_step` | Number | The RecurringStep of the EventBus Schedule. | 何を指すか、範囲が不明 |
| `recurring_unit` | String | The RecurringUnit of the EventBus Schedule. | 取りうる値の列挙がない（例は `day` のみ） |

## sakura_eventbus_trigger

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `process_configuration_id` | String | The ProcessConfiguration ID of the EventBus Trigger. | ID 形式が不明 |
| `source` | String | The source of the EventBus Trigger. | 取りうる値の列挙や例がない |
| `types` | Set of String | The types of the EventBus Trigger. | 取りうる値の列挙や例がない |
| `conditions.op` | String | The operator of the condition for EventBus Trigger. | `eq`/`in` 以外に何があるか不明 |
