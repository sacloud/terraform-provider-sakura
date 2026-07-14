# AppRun / コンテナ / ミドルウェア リソースの不明瞭フィールド詳細

## sakura_apprun_dedicated_application

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `cluster_id` | String | The ID of the cluster. | ID の形式（UUID / 数値 / その他）が不明。 |
| `active_version` | Number | The active version of the application | 数値の範囲や step が不明。 |

## sakura_apprun_dedicated_auto_scaling_group

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `cluster_id` | String | The ID of the cluster. | ID の形式が不明。 |
| `interfaces` | Attributes Set | The network interfaces for the nodes | nested block 全体の説明がなく、各フィールドの意味・依存関係が不明。 |
| `max_nodes` | Number | Maximum number of nodes | 数値の範囲・step が不明。 |
| `min_nodes` | Number | Minimum number of nodes | 数値の範囲・step が不明。 |
| `worker_service_class_path` | String | The worker service class path | path の形式や具体例が不明。 |
| `name_servers` | List of String | The name servers for the auto scaling group (ORDER MATTERS) | IP アドレスの書式例がない。 |
| `interfaces.interface_index` | Number | The interface index | 数値の範囲・step が不明。 |
| `interfaces.upstream` | String | The upstream switch id, or `shared` to use shared segment | switch id の形式が不明。 |
| `interfaces.default_gateway` | String | The default gateway. | IP アドレスの書式例がない。 |
| `interfaces.ip_pool` | Attributes Set | The IP pool for the interface. | nested block の各フィールドの関係が不明。 |
| `interfaces.netmask` | Number | The netmask length. | 範囲・step が不明。 |
| `interfaces.packet_filter_id` | String | The packet filter ID | ID の形式が不明。 |

## sakura_apprun_dedicated_certificate

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `cluster_id` | String | The ID of the cluster. | ID の形式が不明。 |
| `certificate_pem` | String | The PEM-encoded certificate | PEM ファイルの具体例（`BEGIN CERTIFICATE` 等）や書式がない。 |
| `private_key_pem` | String | The PEM-encoded private key | PEM ファイルの具体例や書式がない。 |
| `intermediate_certificate_pem` | String | The PEM-encoded intermediate certificate | PEM ファイルの具体例や書式がない。 |

## sakura_apprun_dedicated_cluster

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `service_principal_id` | String | The service principal ID. This is the principal that invokes the application | ID の形式が不明。 |

## sakura_apprun_dedicated_lb

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `auto_scaling_group_id` | String | The ID of the auto scaling group. | ID の形式が不明。 |
| `cluster_id` | String | The ID of the cluster. | ID の形式が不明。 |
| `interfaces` | Attributes Set | The network interfaces for the load balancer | nested block 全体の説明がなく、各フィールドの意味・依存関係が不明。 |
| `service_class_path` | String | The service class path for the load balancer | path の形式や具体例が不明。 |
| `name_servers` | List of String | The name servers for the load balancer (ORDER MATTERS) | IP アドレスの書式例がない。 |
| `interfaces.interface_index` | Number | The interface index | 数値の範囲・step が不明。 |
| `interfaces.upstream` | String | The upstream switch id, or `shared` to use shared segment | switch id の形式が不明。 |
| `interfaces.default_gateway` | String | The default gateway. | IP アドレスの書式例がない。 |
| `interfaces.ip_pool` | Attributes Set | The IP pool for the interface. | nested block の各フィールドの関係が不明。 |
| `interfaces.netmask` | Number | The netmask length. | 範囲・step が不明。 |
| `interfaces.packet_filter_id` | String | The packet filter ID | ID の形式が不明。 |
| `interfaces.vip` | String | The VIP address. | IP アドレスの書式例がない。 |
| `interfaces.virtual_router_id` | Number | The virtual router ID. | 数値の範囲・step が不明。 |

## sakura_apprun_dedicated_version

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `application_id` | String | The ID of the application. | ID の形式が不明。 |
| `image` | String | The container image | 形式の例（`nginx:latest` 等）が Description にない。 |
| `fixed_scale` | Number | Number of nodes when scaling mode is `manual` | 数値の範囲・step が不明。 |
| `max_scale` | Number | Maximum number of nodes when scaling mode is `autoscale` | 数値の範囲・step が不明。 |
| `min_scale` | Number | Minimum number of nodes when scaling mode is `autoscale` | 数値の範囲・step が不明。 |
| `registry_password_action` | String | Password configuration method | 取りうる値の一覧（enum）がない。 |
| `scale_in_threshold` | Number | When to scale in when scaling mode is `autoscale` | 単位・範囲・step が不明。 |
| `scale_out_threshold` | Number | When to scale out when scaling mode is `autoscale` | 単位・範囲・step が不明。 |
| `exposed_ports.health_check.path` | String | Health check endpoint | path の具体例（`/` 等）が Description にない。 |

## sakura_apprun_shared

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `max_scale` | Number | The maximum number of scales for the entire AppRun Shared application | 数値の範囲・step が不明。 |
| `min_scale` | Number | The minimum number of scales for the entire AppRun Shared application | 数値の範囲・step が不明。 |
| `packet_filter` | Attributes | The packet filter for the AppRun Shared application | nested block の各フィールドの関係が不明。 |
| `components.deploy_source.container_registry.image` | String | The container image name | 形式の例が Description にない。 |
| `components.deploy_source.container_registry.server` | String | The container registry server name | 形式の例が Description にない。 |

## sakura_container_registry

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the Container Registry | ID の形式が不明。 |
| `virtual_domain` | String | The alias for accessing the Container Registry | 値の例や形式がない。 |

## sakura_database

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `network_interface` | Attributes | （Description なし） | nested block 全体の説明がなく、各フィールドの意味・依存関係が不明。 |
| `database_version` | String | The version of the database | 形式の例（`10.11` 等）が Description にない。 |
| `icon_id` | String | The icon id to attach to the Database | ID の形式が不明。 |
| `kms_key_id`（`disk.kms_key_id`） | String | ID of the KMS key for encryption | ID の形式が不明。 |
| `parameters` | Map of String | The map for setting RDBMS-specific parameters. | 設定可能な key / value の具体例がない。 |
| `network_interface.vswitch_id` | String | The id of the vSwitch to which the Database connects | ID の形式が不明。 |
| `network_interface.source_ranges` | List of String | The range of source IP addresses that allow to access to the Database via network | CIDR 表記の書式例がない。 |

## sakura_database_read_replica

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `master_id` | String | The id of the replication master database. | ID の形式が不明。 |
| `network_interface` | Attributes | （Description なし） | nested block 全体の説明がなく、各フィールドの意味・依存関係が不明。 |
| `icon_id` | String | The icon id to attach to the Database Read Replica | ID の形式が不明。 |
| `disk.kms_key_id` | String | ID of the KMS key for encryption | ID の形式が不明。 |
| `network_interface.vswitch_id` | String | The id of the vSwitch to which the Database Replica connects. | ID の形式が不明。 |
| `network_interface.source_ranges` | List of String | The range of source IP addresses that allow to access to the Database Replica via network | CIDR 表記の書式例がない。 |

## sakura_enhanced_db

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `allowed_networks` | List of String | A list of CIDR blocks allowed to connect | CIDR 表記の書式例がない。 |
| `icon_id` | String | The icon id to attach to the Enhanced Database | ID の形式が不明。 |

## sakura_nfs

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `network_interface` | Attributes | The network interface of the NFS. | nested block の内部構造・各フィールドの関係が不明。 |
| `icon_id` | String | The icon id to attach to the NFS | ID の形式が不明。 |
| `network_interface.vswitch_id` | String | The id of the vSwitch to which the NFS connects | ID の形式が不明。 |

## sakura_nosql

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `vswitch_id` | String | The ID of the vSwitch to connect to the NoSQL appliance. | ID の形式が不明。 |
| `remark` | Attributes | （Description なし） | nested block 全体の説明がなく、各フィールドの意味・依存関係が不明。 |
| `settings` | Attributes | Settings of the NoSQL appliance | nested block の各フィールドの意味・依存関係が不明。 |
| `disk.kms_key_id` | String | ID of the KMS key for encryption | ID の形式が不明。 |
| `parameters` | Map of String | Parameters for the NoSQL appliance | 設定可能な key / value の具体例がない。 |
| `remark.nosql.version` | String | Version of database engine used by NoSQL appliance. | 形式の例がない。 |
| `settings.source_network` | List of String | Source network address | CIDR 表記の書式例がない。 |
| `settings.backup.rotate` | Number | Number of backup rotations | 数値の範囲・step が不明。 |

## sakura_nosql_additional_nodes

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `primary_node_id` | String | The ID of the primary node of NoSQL appliance. | ID の形式が不明。 |
| `vswitch_id` | String | The ID of the vSwitch to connect to the Additional nodes of NoSQL appliance. | ID の形式が不明。 |
| `remark` | Attributes | （Description なし） | nested block 全体の説明がなく、各フィールドの意味・依存関係が不明。 |
| `settings` | Attributes | Settings of the Additional nodes of NoSQL appliance | nested block の各フィールドの意味・依存関係が不明。 |

## sakura_ondemand_db

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `allowed_networks` | List of String | A list of CIDR blocks allowed to connect | CIDR 表記の書式例がない。 |
| `icon_id` | String | The icon id to attach to the OnDemand Database | ID の形式が不明。 |

## sakura_simple_mq

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the SimpleMQ | ID の形式が不明。 |
