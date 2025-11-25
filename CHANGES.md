## v3での変更点

### terraform-plugin-frameworkによる書き換え

現在リリースされているv2はSDK v2を利用していますが、v3はFrameworkを利用しています。
muxなども使っておらず、完全移行となります。

### 命名の変更

- `sakuracloud_`プレフィックスは`sakura_`となります。環境変数などを設定することで過去のプレフィックスもサポートする予定です。
- 必要なものはリソース名が適切なものに変更される可能性があります。以下は予定のものになり、他にも増える可能性があります
  - `switch` -> `vswitch`
  - `vpc_router` -> `vpn_router`
  - `proxylb` -> `enhanced_lb`
  - `note` -> `script`
  - `internet` -> `???` (より良い名前を模索中)

### 使われてない機能の削除

#### データソースのfilter

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

### 変更されたリソース

Frameworkでは既存のBlock構文は非推奨になっており、Attribute構文を推奨しています。v3からは過去Block構文を利用していたフィールド群は書き換える必要があります。

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

#### プロバイダの設定

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

#### タイムアウト設定

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

#### container_registry

`user`フィールドがBlockからSet型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
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

```
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

#### dns

`record`フィールドがBlockからLst型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
record {
  name  = "www"
  type  = "A"
  value = "192.168.11.1"
}
record {
  name  = "foobar-dev"
  type  = "CNAME"
  value = "dev.foobar.org"
}
```

- v3

```
record = [
  {
    name  = "www"
    type  = "A"
    value = "192.168.11.1"
  },
  {
    name  = "foobar-dev"
    type  = "CNAME"
    value = "dev.foobar.org"
  }
]
```

#### internet

`assigned_tags`というフィールドが増えています。これは`band_width`の変更によって`id`が変更された場合に自動で付与される`@previous-id`というタグが格納されるフィールドです。v2では`tags`に格納されていましたが、厳格にチェックされるFrameworkによるv3では実現が不可能なため、分離されました。

- v2

```
tags = [
	"tag1",
	"@previous-id=123456789012",
]
```

- v3

```
tags = [
	"tag1",
]
assigned_tags = [
	"@previous-id=123456789012",
]
```

#### nfs

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
  switch_id  = sakura_switch.foobar.id
  ip_address = "192.168.11.101"
  netmask    = 24
  gateway    = "192.168.11.1"
}
```

#### packet_filter / packet_filter_rules

`expression`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
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

```
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

#### server

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

#### vpn_router

`public_network_interface`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
public_network_interface {
  switch_id = sakura_internet.foobar.switch_id
  // ...
}
```

- v3

```
public_network_interface = {
  switch_id = sakura_internet.foobar.switch_id
  // ...
}
```

`private_network_interface`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
private_network_interface {
  index = 1
  // ...
}
```

- v3

```
private_network_interface = [{
  index = 1
  // ...
}]
```

`dhcp_server`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
dhcp_server {
  interface_index = 1
  // ...
}
```

- v3

```
dhcp_server = [{
  interface_index = 1
  // ...
}]
```

`dhcp_static_mapping`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
dhcp_static_mapping {
  ip_address = "192.168.11.10"
  // ...
}
```

- v3

```
dhcp_static_mapping = [{
  ip_address = "192.168.11.10"
  // ...
}]
```

`dns_forwarding`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
dns_forwarding {
  interface_index = 1
  // ...
}
```

- v3

```
dns_forwarding = {
  interface_index = 1
  // ...
}
```

`firewall`フィールドとその中の`expression`がBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
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

```
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

```
l2tp {
  pre_shared_secret = "example"
  // ...
}
```

- v3

```
l2tp = {
  pre_shared_secret = "example"
  // ...
}
```

`pptp`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
pptp {
  range_start = "192.168.11.31"
  // ...
}
```

- v3

```
pptp = {
  range_start = "192.168.11.31"
  // ...
}
```

`port_forwarding`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
port_forwarding {
  protocol = "udp"
  // ...
}
```

- v3

```
port_forwarding = [{
  protocol = "udp"
  // ...
}]
```

`wire_guard`フィールドがBlockからSingle型のAttributeに変更、その中の`peer`がBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
wire_guard {
  ip_address = "192.168.31.1/24"
  peer {
    name = "example"
    // ...
  }
}
```

- v3

```
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

```
site_to_site_vpn {
  peer = "10.0.0.1"
  // ...
}
```

- v3

```
site_to_site_vpn = [{
  peer = "10.0.0.1"
  // ...
}]
```

`site_to_site_vpn_parameter`フィールドとその中の全てのネストされたフィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
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

```
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

```
static_nat {
  public_ip = sakura_internet.foobar.ip_addresses[3]
  // ...
}
```

- v3

```
static_nat = [{
  public_ip = sakura_internet.foobar.ip_addresses[3]
  // ...
}]
```

`static_route`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
static_route {
  prefix = "172.16.0.0/16"
  // ...
}
```

- v3

```
static_route = [{
  prefix = "172.16.0.0/16"
  // ...
}]
```

user`フィールドがBlockからList型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
user {
  name = "username"
  // ...
}
```

- v3

```
user = [{
  name = "username"
  // ...
}]
```

`scheduled_maintenance`フィールドがBlockからSingle型のAttributeに変更されたため、下記のように書き換える必要があります。

- v2

```
scheduled_maintenance {
  day_of_week = "tue"
  // ...
}
```

- v3

```
scheduled_maintenance = {
  day_of_week = "tue"
  // ...
}
```
