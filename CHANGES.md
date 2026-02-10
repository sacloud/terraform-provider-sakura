# v3での変更点

## terraform-plugin-frameworkによる書き換え

現在リリースされているv2はSDK v2を利用していますが、v3はFrameworkを利用しています。
muxなどの互換レイヤーも使っておらず、完全移行となります。

## 命名の変更

- `sakuracloud_`プレフィックスは`sakura_`となります。
- 必要なものはリソース名が適切なものに変更されています。他にも増える可能性があります。
  - `switch` -> `vswitch`
  - `vpc_router` -> `vpn_router`
  - `proxylb` -> `enhanced_lb`
  - `note` -> `script`
  - `load_balancer` -> `dsr_lb`
  - `internet` -> `???` (より良い名前を模索中)
- フィールド名も適切なものに変更されています。
  - `switch_id` -> `vswitch_id` (`local_router`リソースに関してはvswitch以外も対象となるためswitchのまま)
  - `weekdays` -> `days_of_week`

## 使われてない機能の削除

### データソースのfilter

データソースの`filter`ブロックによって実装されていた検索機能は、通常のフィールドでの検索で置き換えられました。例えば以下のようになります。

- v2

```
data "sakuracloud_xxx" "foobar" {
  filter {
    id = "xxxxxxxxxxxx"
    names = ["foobar"]
    tags = ["foo", "bar"]
  }
}
```

- v3

```
data "sakura_xxx" "foobar" {
  id = "xxxxxxxxxxxx"
  name = "foobar" // namesからnameに変わっている
  tags = ["foo", "bar"]
}
```

## 変更されたリソース

Frameworkではフィールドを指定するためにBlock構文を使用することは非推奨になっており、Attribute構文を推奨しています。v3からは過去Block構文を利用していたフィールド群は書き換える必要があります。

```
# Attribute構文。リストやブロックでもこちらで書くのを推奨されている
user = [
  {
    //...
  },
  {
    //...
  }
]

# Block構文。こちらは古い書き方で、現状Block機能を使うことで互換性のために実装できるが非推奨
user {
  // ...
}
user {
  // ...
}
```

また、Frameworkはより厳密に型や値をチェックするようになったため、SDK v2でチェックされない挙動に依存してたリソースも一部挙動が変更されています。

### プロバイダの設定

`sakuracloud`を`sakura`に書き換えてください。

- v2

```
terraform {
  required_providers {
    sakuracloud = {
      source = "sacloud/sakuracloud"
    }
  }
}

provider "sakuracloud" {
  zone = "tk1b"
}
```

- v3

```
terraform {
  required_providers {
    sakura = {
      source = "sacloud/sakura"
    }
  }
}

provider "sakura" {
  zone = "tk1b"
}
```

### タイムアウト設定

BlockからAttributeに変更されたため、以下のように書き換える必要があります

- v2

```
timeouts {
  create = "20m"
}
```

- v3

```
timeouts = {
  create = "20m"
}
```

### apprun_shared

v3.1.0からwrite-only版の`password_wo`/`password_wo_version`が提供されました。今後はこちらを利用してください。

`components`フィールドがBlockからList型Attribute、内部で設定可能なパラメータもAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  components {
    name = "foobar"
    // ...
    deploy_source {
      container_registry {
        image = "foorbar.sakuracr.jp/foorbar:latest"
        // ...
      }
    }
    env {
      key   = "key"
      value = "value"
    }
    probe {
      http_get {
        path = "/"
        // ...
        headers {
          name  = "name"
          value = "value"
        }
      }
    }
  }
```

- v3

```hcl
  components = [{
    name = "foobar"
    // ...
    deploy_source = {
      container_registry = {
        image = "foobar.sakuracr.jp/my-app:latest"
        // ...
      }
    }
    env = [{
      key   = "key"
      value = "value"
    }]
    probe = {
      http_get = {
        path = "/"
        // ...
        headers = [{
          name  = "name"
          value = "value"
        }]
      }
    }
  }]
```

`traffics`フィールドがBlockからList型Attributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  traffics {
    version_index = 0
    percent       = 100
  }
```

- v3

```hcl
  traffics = [{
    version_index = 0
    percent       = 100
  }]
```

`packet_filter`フィールドがBlockからSingle型Attributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  packet_filter {
    enabled = true
    settings {
      from_ip               = "192.0.2.0"
      from_ip_prefix_length = "24"
    }
  }
```


- v3

```hcl
  packet_filter = {
    enabled = true
    settings = [{
      from_ip               = "192.0.2.0"
      from_ip_prefix_length = "24"
    }]
  }
```

### auto_backup

`weekdays`フィールドは`days_of_week`に変更されました。

### auto_scale

`disabled`フィールドは逆の真偽値を指定する`enabled`に変更されました。これは他のリソースとパラメータ名を統一するためです。指定していなかった場合の挙動はかわっていません。

`trigger_type`はオプションではなく必須フィールドとなりました。

`router_threshold_scaling` / `cpu_threshold_scaling` / `schedule_scaling`はBlock型からAttribute型に変更されました。下記のように書き換える必要があります。

- v2

```hcl
cpu_threshold_scaling {
  // ...
}

router_threshold_scaling {
  // ...
}

schedule_scaling {
  // ...
}
schedule_scaling {
  // ...
}
```

- v3

```hcl
cpu_threshold_scaling = {
  // ...
}

router_threshold_scaling = {
  // ...
}

schedule_scaling = [{
  // ...
},
{
  // ...
}]
```

### container_registry

`user`フィールドがBlockからSet型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
user {
  name       = "user1"
  password   = "user1_pass"
  permission = "readonly"
}
user {
  name       = "user2"
  password   = "user2_pass"
  permission = "all"
}
```

- v3

```hcl
user = [
  {
    name       = "user1"
    password   = "user1_pass"
    permission = "readonly"
  },
  {
    name       = "user2"
    password   = "user2_pass"
    permission = "all"
  }
]
```

### database

v3.1.0からwrite-only版の`password_wo`/`password_wo_version`が提供されました。今後はこちらを利用してください。

`switch_id`フィールドは`vswitch_id`に変更されました。
`weekdays`フィールドは`days_of_week`に変更されました。

`network_interface`フィールドがBlockからSingle型Attributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  network_interface {
    switch_id  = sakuracloud_switch.foobar.id
    ip_address = "192.168.11.11"
    // ...
  }
```

- v3

```hcl
  network_interface = {
    vswitch_id = sakura_vswitch.foobar.id
    ip_address = "192.168.11.11"
    // ...
  }
```

`backup`フィールドがBlockからSingle型Attributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
 backup {
    time     = "00:00"
    weekdays = ["mon", "tue"]
  }
```

- v3

```hcl
  backup = {
    time     = "00:00"
    weekdays = ["mon", "tue"]
  }
```

### database_read_replica

`switch_id`フィールドは`vswitch_id`に変更されました。

`network_interface`フィールドがBlockからSingle型Attributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  network_interface {
    ip_address = "192.168.11.111"
    // ...
  }
```

- v3

```hcl
  network_interface = {
    ip_address = "192.168.11.111"
    // ...
  }
```

### dns

`record`フィールドは削除されました。レコードを設定するには`dns_record`リソースを利用してください。

- v2

```hcl
# dnsリソースでもレコードを設定可能だった
resource "sakuracloud_dns" "foobar" {
  zone = "foobar.example.com"
  record = [{
    name  = "www"
    type  = "A"
    value = "192.168.11.1"
  }]
}

resource "sakuracloud_dns_record" "record" {
  dns_id = sakura_dns.foobar.id
  name  = "www"
  type  = "A"
  value = "192.168.11.2"
}

```

- v3

```hcl
# dnsリソースではレコードは設定不可
resource "sakura_dns" "foobar" {
  zone = "foobar.example.com"
}

resource "sakura_dns_record" "record1" {
  dns_id = sakura_dns.foobar.id
  name  = "www"
  type  = "A"
  value = "192.168.11.1"
}

resource "sakura_dns_record" "record2" {
  dns_id = sakura_dns.foobar.id
  name  = "www"
  type  = "A"
  value = "192.168.11.2"
}
```

これは複数のリソースから同一のデータを更新することによるトラブルを避けるための変更となります。

### dsr_lb(load_balancer in v2)

`switch_id`フィールドは`vswitch_id`に変更されました。

`network_interface` フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  network_interface {
    vrid = 1
    // ...
  }
```

- v3

```hcl
  network_interface = {
    vrid = 1
    // ...
  }
```

`vid` / `server`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  vip {
    vip = "192.168.11.201"
    // ...

    server {
      ip_address = "192.168.11.51"
      // ...
    }
    server {
      ip_address = "192.168.11.52"
      // ...
    }
  }
```

- v3

```hcl
  vip = [{
    vip = "192.168.11.201"
    // ...

    server = [{
      ip_address = "192.168.11.51"
      // ...
    },
    {
      ip_address = "192.168.11.52"
      // ...
    }]
  }]
```

### enhanced_db

`password`からwrite-only版の`password_wo`/`password_wo_version`に変更されました。

### enhanced_lb(proxylb in v2)

`health_check` / `sorry_server` / `syslog`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  health_check {
    protocol    = "http"
    // ...
  }
  sorry_server {
    ip_address = "192.0.2.1"
    port       = 80
  }
  syslog {
    server = "192.0.2.1"
    port   = 514
  }
```

- v3

```hcl
  health_check = {
    protocol    = "http"
    // ...
  }
  sorry_server = {
    ip_address = "192.0.2.1"
    port       = 80
  }
  syslog = {
    server = "192.0.2.1"
    port   = 514
  }
```

`bind_port` / `server` / `rule`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  bind_port {
    proxy_mode = "http"
    port       = 80
    response_header {
      // ...
    }
  }

  server = {
    ip_address = sakuracloud_server.foobar.ip_address
    // ...
  }

  rule {
    action = "forward"
    // ...
  }
  rule {
    action = "redirect"
    // ...
  }
```

- v3

```hcl
  bind_port = [{
    proxy_mode = "http"
    port       = 80
    response_header = [{
      // ...
    }]
  }]

  server = [{
    ip_address = sakura_server.foobar.ip_address
    // ...
  }]

  rule = [{
    action = "forward"
    // ...
  },
  {
    action = "redirect"
    // ...
  }]
```

### gslb

`health_check`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  health_check {
    protocol = "http"
    // ...
  }
```

- v3

```hcl
  health_check = {
    protocol = "http"
    // ...
  }
```

`server`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  server {
    ip_address = "192.2.0.11"
    weight     = 1
    enabled    = true
  }
```

- v3

```hcl
  server = [{
    ip_address = "192.2.0.11"
    weight     = 1
    enabled    = true
  }]
```

### internet

`switch_id`フィールドは`vswitch_id`に変更されました。

### local_router

`switch` / `network_interface`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  switch {
    code = sakuracloud_switch.foobar.id
    // ....
  }

  network_interface {
    vip = "192.168.11.1"
    // ...
  }
```

- v3

```hcl
  switch = {
    code = sakura_vswitch.foobar.id
    // ....
  }

  network_interface = {
    vip = "192.168.11.1"
    // ...
  }
```


`static_route` / `peer`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
  static_route {
    prefix   = "10.0.0.0/24"
    next_hop = "192.168.11.2"
  }
  static_route {
    prefix   = "172.16.0.0/16"
    next_hop = "192.168.11.3"
  }

  peer {
    peer_id = data.sakuracloud_local_router.peer.id
    // ...
  }
```

- v3

```hcl
  static_route = [{
    prefix   = "10.0.0.0/24"
    next_hop = "192.168.11.2"
  },
  {
    prefix   = "172.16.0.0/16"
    next_hop = "192.168.11.3"
  }]

  peer = [{
    peer_id = data.sakura_local_router.peer.id
  }]
```

### nfs

`switch_id`フィールドは`vswitch_id`に変更されました。

`network_interface`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
network_interface {
  switch_id  = sakuracloud_switch.foobar.id
  ip_address = "192.168.11.101"
  netmask    = 24
  gateway    = "192.168.11.1"
}
```

- v3

```
network_interface = {
  vswitch_id = sakura_vswitch.foobar.id
  ip_address = "192.168.11.101"
  netmask    = 24
  gateway    = "192.168.11.1"
}
```

### packet_filter / 

`expression`フィールドが削除されました。ルールを設定するには`packet_filter_rules`リソースを利用してください。

- v2

```hcl
resource "sakuracloud_packet_filter" "foobar" {
  name        = "foobar"
  description = "description"
  expression {
    protocol         = "tcp"
    destination_port = "22"
  }
}
```

- v3

```hcl
resource "sakura_packet_filter" "foobar" {
  name        = "foobar"
  description = "description"
}

resource "sakura_packet_filter_rules" "rules" {
  packet_filter_id = sakura_packet_filter.foobar.id

  expression = [{
    protocol         = "tcp"
    destination_port = "22"
  }]
}
```

### packet_filter_rules

`expression`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
expression {
  protocol         = "tcp"
  destination_port = "22"
}
expression {
  protocol    = "udp"
  source_port = "123"
}
```

- v3

```hcl
expression = [
  {
    protocol         = "tcp"
    destination_port = "22"
  },
  {
    protocol    = "udp"
    source_port = "123"
  }
]
```

### server

v3.1.0からwrite-only版の`password_wo`/`password_wo_version`が提供されました。今後はこちらを利用してください。

`disk_edit_parameter`内の`note_ids`フィールドが削除されました。代わりに`disk_edit_parameter`内のList型の`script`フィールドを利用してください。

`network_interface`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
network_interface {
  upstream         = "shared"
  packet_filter_id = data.sakuracloud_packet_filter.foobar.id
}
network_interface {
  // その他の設定
}
```

- v3

```
network_interface = [
  {
    upstream         = "shared"
    packet_filter_id = data.sakura_packet_filter.foobar.id
  },
  {
    // その他の設定
  }
]
```

`disk_edit_parameter`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
disk_edit_parameter {
  hostname = "foobar"
  password = "foobar-password"
  ssh_key_ids = ["xxxxxxxxxxxx"]
  // ...
}
```

- v3

```hcl
disk_edit_parameter = {
  hostname = "foobar"
  password = "foobar-password"
  ssh_key_ids = ["xxxxxxxxxxxx"]
  // ...
}
```

### simple_monitor

v3.1.0からwrite-only版の`password_wo`/`password_wo_version`が提供されました。今後はこちらを利用してください。

`health_check`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
health_check {
  protocol = "https"
  // ...
}
```

- v3

```hcl
health_check = {
  protocol = "https"
  // ...
}
```

### subnet

`switch_id`フィールドは`vswitch_id`に変更されました。

### vpn_router

v3.1.0からwrite-only版の`password_wo`/`password_wo_version`が提供されました。今後はこちらを利用してください。

`switch_id`フィールドは`vswitch_id`に変更されました。

`public_network_interface`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
public_network_interface {
  switch_id = sakura_internet.foobar.switch_id
  // ...
}
```

- v3

```hcl
public_network_interface = {
  vswitch_id = sakura_internet.foobar.vswitch_id
  // ...
}
```

`private_network_interface`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
private_network_interface {
  index = 1
  // ...
}
```

- v3

```hcl
private_network_interface = [{
  index = 1
  // ...
}]
```

`dhcp_server`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
dhcp_server {
  interface_index = 1
  // ...
}
```

- v3

```hcl
dhcp_server = [{
  interface_index = 1
  // ...
}]
```

`dhcp_static_mapping`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
dhcp_static_mapping {
  ip_address = "192.168.11.10"
  // ...
}
```

- v3

```hcl
dhcp_static_mapping = [{
  ip_address = "192.168.11.10"
  // ...
}]
```

`dns_forwarding`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
dns_forwarding {
  interface_index = 1
  // ...
}
```

- v3

```hcl
dns_forwarding = {
  interface_index = 1
  // ...
}
```

`firewall`フィールドとその中の`expression`がBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
firewall {
  interface_index = 1
  direction = "send"
  expression {
    protocol = "tcp"
    // ...
  }
  // ...
}
```

- v3

```hcl
firewall = [{
  interface_index = 1
  direction = "send"
  expression = [
    {
      protocol = "tcp"
      // ...
    },
    // ...
  ]
}
```

`l2tp`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
l2tp {
  pre_shared_secret = "example"
  // ...
}
```

- v3

```hcl
l2tp = {
  pre_shared_secret = "example"
  // ...
}
```

`pptp`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
pptp {
  range_start = "192.168.11.31"
  // ...
}
```

- v3

```hcl
pptp = {
  range_start = "192.168.11.31"
  // ...
}
```

`port_forwarding`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
port_forwarding {
  protocol = "udp"
  // ...
}
```

- v3

```hcl
port_forwarding = [{
  protocol = "udp"
  // ...
}]
```

`wire_guard`フィールドがBlockからSingle型のAttributeに変更、その中の`peer`がBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
wire_guard {
  ip_address = "192.168.31.1/24"
  peer {
    name = "example"
    // ...
  }
}
```

- v3

```hcl
wire_guard = {
  ip_address = "192.168.31.1/24"
  peer = [{
    name = "example"
    // ...
  }]
}
```

`site_to_site_vpn`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
site_to_site_vpn {
  peer = "10.0.0.1"
  // ...
}
```

- v3

```hcl
site_to_site_vpn = [{
  peer = "10.0.0.1"
  // ...
}]
```

`site_to_site_vpn_parameter`フィールドとその中の全てのネストされたフィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
site_to_site_vpn_parameter {
  ike {
    lifetime = 28800
    dpd {
      interval = 15
      timeout  = 30
    }
  }
  esp {
    lifetime = 1800
  }
  encryption_algo = "aes256"
  // ...
}
```

- v3

```hcl
site_to_site_vpn_parameter = {
  ike = {
    lifetime = 28800
    dpd = {
      interval = 15
      timeout  = 30
    }
  }
  esp = {
    lifetime = 1800
  }
  encryption_algo = "aes256"
  // ...
}
```

`static_nat`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
static_nat {
  public_ip = sakura_internet.foobar.ip_addresses[3]
  // ...
}
```

- v3

```hcl
static_nat = [{
  public_ip = sakura_internet.foobar.ip_addresses[3]
  // ...
}]
```

`static_route`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
static_route {
  prefix = "172.16.0.0/16"
  // ...
}
```

- v3

```hcl
static_route = [{
  prefix = "172.16.0.0/16"
  // ...
}]
```

user`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
user {
  name = "username"
  // ...
}
```

- v3

```hcl
user = [{
  name = "username"
  // ...
}]
```

`scheduled_maintenance`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```hcl
scheduled_maintenance {
  day_of_week = "tue"
  // ...
}
```

- v3

```hcl
scheduled_maintenance = {
  day_of_week = "tue"
  // ...
}
```
