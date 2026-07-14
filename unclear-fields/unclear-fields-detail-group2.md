# ネットワーク リソースの不明瞭フィールド詳細

## sakura_internet

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the Internet(switch+router) | ID の形式（UUID / 数値 等）が記載されていない |

## sakura_vswitch

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `bridge_id` | String | The bridge id attached to the vSwitch | どのリソースの ID か、形式が不明 |
| `icon_id` | String | The icon id to attach to the vSwitch | ID の形式が不明 |

## sakura_switch

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `bridge_id` | String | The bridge id attached to the switch | どのリソースの ID か、形式が不明 |
| `icon_id` | String | The icon id to attach to the Switch | ID の形式が不明 |

## sakura_subnet

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `internet_id` | String | The id of the Internet(switch+router) resource that the Subnet belongs | リソース ID であることは示されているが、UUID か数値か等の形式が不明 |

## sakura_bridge

問題なし

## sakura_local_router

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the Local Router | ID の形式が不明 |
| `network_interface.ip_addresses` | List of String | The list of the IP address assigned | いくつ IP が必要か、VIP との関係 / CIDR 表記の可否が不明 |
| `network_interface.netmask` | Number | The bit length of the subnet assigned to the network interface | 取值範囲が記載されていない |
| `network_interface.vip` | String | The virtual IP address | IP 書式の例がない |
| `network_interface.vrid` | Number | The Virtual Router Identifier | 取值範囲や形式の例がない |
| `switch.code` | String | The resource ID of the Switch | ID の形式が不明 |
| `switch.zone` | String | The name of the Zone | ゾーン名の例（`is1a` 等）がない |
| `peer.peer_id` | String | The ID of the peer LocalRouter | ID の形式が不明 |
| `peer.enabled` | Boolean | The flag to enable the LocalRouter | 「LocalRouter 自身」を有効化するのか「ピア設定」を有効化するのか曖昧 |

## sakura_gslb

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the GSLB | ID の形式が不明 |
| `health_check.port` | Number | The port number used when checking by TCP/HTTP/HTTPS | ポート番号の取值範囲が記載されていない |
| `health_check.status` | String | The response-code to expect when checking by HTTP/HTTPS | 期待するステータスコードの例（`200` 等）がない |

## sakura_enhanced_lb

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the Enhanced LB | ID の形式が不明 |
| `timeout` | Number | The timeout duration in seconds | 取值範囲や step が記載されていない |
| `bind_port.port` | Number | The number of listening port | ポート番号の取值範囲が記載されていない |
| `sorry_server.port` | Number | The port number of the SorryServer | ポート番号の取值範囲が記載されていない |
| `syslog.port` | Number | The number of syslog port | ポート番号の取值範囲が記載されていない |

## sakura_enhanced_lb_acme

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `enhanced_lb_id` | String | The id of the Enhanced LB that set ACME settings to | ID の形式が不明 |
| `get_certificates_timeout_sec` | Number | The timeout in seconds for the certificate acquisition to complete | 取值範囲が記載されていない |
| `update_delay_sec` | Number | The wait time in seconds. This typically used for waiting for a DNS propagation | 取值範囲や推奨値が記載されていない |
| `subject_alt_names` | Set of String | The Subject alternative names used by ACME | 入力すべき FQDN / ワイルドカード等の書式例がない |

## sakura_dsr_lb

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the DSR LB | ID の形式が不明 |
| `network_interface.ip_addresses` | List of String | A list of IP address to assign to the DSR LB | いくつ IP が必要か、VIP との関係が不明 |
| `network_interface.vrid` | Number | The Virtual Router Identifier | 取值範囲や形式の例がない |
| `network_interface.vswitch_id` | String | The id of the vSwitch to which the DSR LB connects | ID の形式が不明 |
| `vip.server.connect_timeout` | Number | The timeout in seconds for health checks, available only for TCP/HTTP/HTTPS | 取值範囲が記載されていない |
| `vip.server.path` | String | The path used when checking by HTTP/HTTPS | パスの書式例がない |
| `vip.server.retry` | Number | The retry count for server down detection, available only for TCP/HTTP/HTTPS | 取值範囲が記載されていない |
| `vip.server.status` | Number | The response code to expect when checking by HTTP/HTTPS | ステータスコードの例がない |

## sakura_vpn_router

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the VPN Router | ID の形式が不明 |
| `version` | Number | The version of the VPN Router | 取值範囲や指定可能な値が不明 |
| `dhcp_static_mapping.mac_address` | String | The source MAC address of static mapping | MAC アドレスの書式例（`aa:bb:cc:aa:bb:cc` 等）がない |
| `l2tp.pre_shared_secret` | String, Sensitive | The pre shared secret for L2TP/IPsec | 文字数制限や推奨形式が記載されていない |
| `l2tp.pre_shared_secret_wo` | String | The pre shared secret for L2TP/IPsec | 文字数制限や推奨形式が記載されていない |
| `port_forwarding.private_port` | Number | The destination port number of the port forwarding. This will be a port number on a private network | ポート番号の取值範囲が記載されていない |
| `port_forwarding.public_port` | Number | The source port number of the port forwarding. This must be a port number on a public network | ポート番号の取值範囲が記載されていない |
| `private_network_interface.netmask` | Number | The bit length of the subnet assigned to the network interface | 取值範囲が記載されていない |
| `private_network_interface.vswitch_id` | String | The id of the connected vSwitch | ID の形式が不明 |
| `public_network_interface.vrid` | Number | The Virtual Router Identifier. This is only required when `plan` is not `standard` | 取值範囲や形式の例がない |
| `public_network_interface.vswitch_id` | String | The id of the vSwitch to connect. This is only required when `plan` is not `standard` | ID の形式が不明 |
| `scheduled_maintenance.hour` | Number | The time to start maintenance | 取值範囲（0〜23 等）や単位が不明 |
| `site_to_site_vpn.remote_id` | String | The id of the opposing appliance connected to the VPN Router. This is typically set same as value of `peer` | ID の形式が不明 |
| `site_to_site_vpn_parameter.esp.lifetime` | Number | Default: 1800 | 取值範囲や単位（秒）以外の制約が不明 |
| `site_to_site_vpn_parameter.ike.lifetime` | Number | Lifetime of IKE SA. Default: 28800 | 取值範囲が記載されていない |
| `site_to_site_vpn_parameter.ike.dpd.interval` | Number | Default: 15 | 取值範囲が記載されていない |
| `site_to_site_vpn_parameter.ike.dpd.timeout` | Number | Default: 30 | 取值範囲が記載されていない |
| `wire_guard.peer.public_key` | String | the public key of the WireGuard client | WireGuard 公開鍵の書式例（Base64 文字列等）がない |

## sakura_packet_filter

問題なし

## sakura_packet_filter_rules

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `packet_filter_id` | String | The id of the packet filter that set expressions to | ID の形式が不明 |
| `expression.description` | String | The description of this packet filter expression | 文字数制限等の制約が記載されていない |

## sakura_ipv4_ptr

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `retry_interval` | Number | The wait interval(in seconds) for retrying API call used when SakuraCloud API returns any errors | 取值範囲や推奨値が記載されていない |
| `retry_max` | Number | The maximum number of API call retries used when SakuraCloud API returns any errors | 取值範囲が記載されていない |

## sakura_dns

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `icon_id` | String | The icon id to attach to the DNS | ID の形式が不明 |

## sakura_dns_record

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `dns_id` | String | The id of the DNS resource | ID の形式が不明 |
| `type` | String | The type of the DNS record | 取りうるレコードタイプの列挙（A/CNAME/TXT 等）がない |
| `value` | String | The value of the DNS record | レコードタイプごとの書式例がない |
| `ttl` | Number | The TTL of the DNS record | 取值範囲や単位（秒）が記載されていない |
