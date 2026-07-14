# コンピュート / ストレージ リソースの不明瞭フィールド詳細

## sakura_server

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `cdrom_id` | String | The id of the CD-ROM to attach to the Server | ID の形式（数値 / UUID 等）が不明 |
| `core` | Number | The number of virtual CPUs | 取りうる範囲、step、マシンスペックとの制約が不明 |
| `cpu_model` | String | The model of cpu | 取りうる値の一覧が不明 |
| `disks` | List of String | A list of disk id connected to the server | ID の形式が不明 |
| `gpu` | Number | The number of GPUs | 取りうる範囲、step、どのプランで利用可能か不明 |
| `gpu_model` | String | The model of gpu | 取りうる値の一覧が不明 |
| `icon_id` | String | The icon id to attach to the Server | ID の形式が不明 |
| `memory` | Number | The size of memory in GiB | 取りうる範囲、step が不明 |
| `private_host_id` | String | The id of the PrivateHost which the Server is assigned | ID の形式が不明 |
| `private_host_name` | String | The id of the PrivateHost which the Server is assigned | 名前フィールドなのに Description が「id」と記載されており、何を入力すべきか分からない |
| `disk_edit_parameter.gateway` | String | The gateway address used by the Server | 書式例（IPv4 アドレス）がない |
| `disk_edit_parameter.netmask` | Number | The bit length of the subnet to assign to the Server | 取りうる範囲（例: 1-32）が不明 |
| `disk_edit_parameter.script.id` | String | The id of the script | ID の形式が不明 |
| `disk_edit_parameter.script.api_key_id` | String | The id of the API key to be injected into script when editing the disk | どの API キー、ID の形式が不明 |
| `disk_edit_parameter.script.variables` | Map of String | The value of the variable that be injected into script when editing the disk | 変数の命名規則や値の例がない |
| `disk_edit_parameter.ssh_key_ids` | Set of String | A set of the SSHKey id | ID の形式が不明 |
| `disk_edit_parameter.ssh_keys` | Set of String | A set of the SSHKey text | 想定される公開鍵形式（OpenSSH / RFC4716 等）が不明 |
| `network_interface.packet_filter_id` | String | The id of the packet filter to attach to the network interface | ID の形式が不明 |

## sakura_disk

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `dedicated_storage_id` | String | ID of the dedicated storage | ID の形式が不明 |
| `distant_from` | Set of String | A list of disk id. The disk will be located to different storage from these disks | ID の形式が不明 |
| `icon_id` | String | The icon id to attach to the Disk | ID の形式が不明 |
| `kms_key_id` | String | ID of the KMS key for encryption | ID の形式が不明 |
| `server_id` | String | The id of the server connected to the Disk | ID の形式が不明 |
| `size` | Number | The size of Disk in GiB | 取りうる範囲、step、プラン依存の制約が不明 |
| `source_archive_id` | String | The id of the source archive. This conflicts with [`source_disk_id`] | ID の形式が不明 |
| `source_disk_id` | String | The id of the source disk. This conflicts with [`source_archive_id`] | ID の形式が不明 |

## sakura_private_host

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `dedicated_storage_id` | String | ID of the dedicated storage (input-only). This value is not returned by the backend API, so Terraform cannot detect drift for this attribute. Note: it cannot be restored via `terraform import` | ID の形式が不明 |
| `icon_id` | String | The icon id to attach to the PrivateHost | ID の形式が不明 |

## sakura_cdrom

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the CD-ROM | ID の形式が不明 |

## sakura_ssh_key

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `public_key` | String | The body of the public key | 想定される公開鍵形式（例: `ssh-rsa AAAAB3...` / `ssh-ed25519 ...` 等）の記載がない |

## sakura_icon

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `base64content` | String | The base64 encoded content to upload as the Icon | どの画像形式（PNG / JPEG / GIF 等）を base64 エンコードすればよいか不明 |
| `source` | String | The file path to upload as the Icon | 対応する画像形式の記載がない |

## sakura_script

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the Script | ID の形式が不明 |

## sakura_archive

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `archive_file` | String | The file path to upload to the SakuraCloud Archive | 対応するファイル形式（raw / qcow2 / vmdk 等）の記載がない |
| `hash` | String | The md5 checksum calculated from the base64 encoded file body | md5 ハッシュの形式（32 文字の16進数等）の記載がない |
| `icon_id` | String | The icon id to attach to the Archive | ID の形式が不明 |
| `source_archive_id` | String | The id of the source archive. This conflicts with [`source_disk_id`] | ID の形式が不明 |
| `source_archive_zone` | String | The share key of source shared archive | ゾーン名なのか共有キーなのか不明（Description とフィールド名が矛盾） |
| `source_disk_id` | String | The id of the source disk. This conflicts with [`source_archive_id`] | ID の形式が不明 |
| `source_shared_key` | String | The share key of source shared archive | 共有キーの形式（例: `is1a:123456789012:xxx`）の記載がない |
| `source_shared_key_wo` | String | The share key of source shared archive | 共有キーの形式の記載がない |

## sakura_dedicated_storage

問題なし

## sakura_auto_scale

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `api_key_id` | String | The ID of the API key | どの API キー（sakura_cloud API キー / sacloud/autoscaler 用等）、ID の形式が不明 |
| `config` | String | The configuration file for sacloud/autoscaler | 設定ファイルの形式（YAML / HCL / JSON 等）やスキーマ、必須項目の記載がない |
| `cpu_threshold_scaling.down` | Number | Threshold for average CPU utilization to scale down/in | 単位（% かどうか）、取りうる範囲が不明 |
| `cpu_threshold_scaling.up` | Number | Threshold for average CPU utilization to scale up/out | 単位（% かどうか）、取りうる範囲が不明 |
| `router_threshold_scaling.mbps` | Number | Mbps | 取りうる範囲、step が不明 |
| `schedule_scaling.hour` | Number | Hour to be triggered | 取りうる範囲（0-23 か等）、書式の記載がない |

## sakura_auto_backup

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `disk_id` | String | The disk id to backed up | ID の形式が不明 |
| `icon_id` | String | The icon id to attach to the AutoBackup | ID の形式が不明 |

## sakura_cloudhsm

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `ipv4_netmask` | Number | The IPv4 netmask of the CloudHSM | 取りうる範囲（例: 1-32）の記載がない |
| `ipv4_network_address` | String | The IPv4 network address of the CloudHSM | ネットワークアドレス（ホストビットが 0）であること、`ipv4_netmask` との関係などの書式例がない |

## sakura_cloudhsm_client

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `certificate` | String | The certificate for the CloudHSM Client | 想定される証明書形式（PEM / DER 等、ファイル内容そのものか）の記載がない |
| `cloudhsm_id` | String | The ID of the CloudHSM to associate with the client | ID の形式が不明 |

## sakura_cloudhsm_license

問題なし

## sakura_cloudhsm_peer

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `cloudhsm_id` | String | The ID of the CloudHSM to associate with the peer | ID の形式が不明 |
| `router_id` | String | The router ID to associate with the peer | どのルータリソースの ID か、ID の形式が不明 |
| `secret_key` | String | The secret key for the CloudHSM Peer. | シークレットキーの形式や長さの記載がない |
| `secret_key_wo` | String | The secret key for the CloudHSM Peer | シークレットキーの形式や長さの記載がない |

## sakura_kms

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `plain_key` | String | Plain key for imported KMS key. Required when `key_origin` is 'imported'. | キーの形式、長さ、エンコーディング（Base64 / 生バイナリ等）の記載がない |
| `plain_key_wo` | String | Plain key for imported KMS key. Required when `key_origin` is 'imported'. | キーの形式、長さ、エンコーディングの記載がない |
| `rotate_version` | Number | The rotation version. This number is incremented when you want rotate KMS key. | 取りうる範囲、step、初期値の記載がない |
| `schedule_destruction_days` | Number | The number of days to schedule the destruction of the KMS key. If set, the KMS key will be scheduled for destruction after the specified number of days instead of immediate destruction in 'terraform destroy'. | 取りうる範囲、最小/最大値の記載がない |
| `status` | String | The status of the KMS key. | 取りうる値の一覧が不明。また Optional だが変更可能か計算値か不明 |

## sakura_secret_manager

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `kms_key_id` | String | KMS key ID for the SecretManager vault. | ID の形式が不明 |

## sakura_secret_manager_secret

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `vault_id` | String | The Secret Manager's vault id. | ID の形式が不明 |
| `value` | String | Secret value. | 値の形式、長さ制限、許容される文字種などの記載がない |
| `value_wo` | String | Secret value. (write-only) | 値の形式、長さ制限、許容される文字種などの記載がない |
