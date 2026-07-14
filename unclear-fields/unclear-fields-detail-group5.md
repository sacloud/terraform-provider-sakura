# WebAccel / Addon / Object Storage / API Gateway リソースの不明瞭フィールド詳細

## sakura_webaccel

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `origin_parameters.origin` | String | Origin hostname or IP address. Required for type = web | IP アドレスの書式（CIDR 表記の有無、IPv6 対応等）、FQDN のみかサブドメインも可か、末尾の `.` 有無等の例がない |
| `origin_parameters.host_header` | String | Host header to the origin. Optional for type = web | どのような文字列を入れるか（ドメイン名、ポート付きか等）の例がない |
| `origin_parameters.protocol` | String | Request protocol for the origin host. Required for type = web | 取りうる値の一覧がない（`http`/`https`/その他？） |
| `origin_parameters.endpoint` | String | Object Storage's S3 endpoint without protocol scheme. Required for type = bucket | 値の例（`s3.isk01.sakurastorage.jp` 等）が Description にない |
| `origin_parameters.region` | String | Object Storage's S3 region. Required for type = bucket | 取りうるリージョン値の一覧がない |
| `origin_parameters.bucket_name` | String | Object Storage's bucket name. Required for type = bucket | 命名規則や使用可能文字、長さ制限等が不明 |
| `origin_parameters.access_key_wo` | String | Object Storage's access key. Required for type = bucket | 値の形式（文字列長、文字種等）が不明 |
| `origin_parameters.secret_access_key_wo` | String | Object Storage's secret access key. Required for type = bucket | 値の形式（文字列長、文字種等）が不明 |
| `origin_parameters.use_document_index` | Boolean | Whether the document indexing for the bucket is enabled or not. Optional for type = bucket | true にした場合の動作や対象インデックスファイル名（`index.html` 等）の説明がない |
| `default_cache_ttl` | Number | The default cache TTL of the site | 単位（秒/分/時間）と範囲、step が不明 |
| `domain` | String | Domain name of the site. Required when domain_type is own_domain | 入力するドメインの形式例（サブドメイン含むか、末尾 `.` の扱い等）がない |
| `logging.endpoint` | String | Object Storage's S3 endpoint without protocol scheme | 値の例が Description にない |
| `logging.region` | String | Object Storage's S3 region | 取りうるリージョン値の一覧がない |
| `logging.bucket_name` | String | Object Storage's bucket name | 命名規則や使用可能文字、長さ制限等が不明 |
| `logging.access_key_wo` | String | Object Storage's access key | 値の形式が不明 |
| `logging.secret_access_key_wo` | String | Object Storage's secret access key | 値の形式が不明 |
| `cors_rules.allowed_origins` | List of String | List of allowed origins for CORS | origin の書式例（`https://example.com`、ワイルドカード等）がない |
| `onetime_url_secrets_wo` | List of String | The site-wide onetime url secrets | 各文字列の形式、長さ、推奨生成方法等が不明 |

## sakura_webaccel_acl

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `site_id` | String | The site ID of WebAccel. | どのリソースの ID か、形式（UUID/数値/その他）が不明 |
| `acl` | String | ACL definition for the WebAccel site. | 構文全体の仕様（`deny`/`allow` 以外のディレクティブ、IP/CIDR の書式、ポート指定の有無等）が不明 |

## sakura_webaccel_activation

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `site_id` | String | The site ID of WebAccel. | どのリソースの ID か、形式（UUID/数値/その他）が不明 |

## sakura_webaccel_certificate

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `site_id` | String | The site ID of WebAccel. | どのリソースの ID か、形式（UUID/数値/その他）が不明 |

## sakura_addon_ai

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `location` | String | The location of the Addon AI | 例は `japaneast` だが、取りうる値の一覧が不明 |
| `sku` | Number | The SKU of the Addon AI | 取りうる値の一覧、範囲、各 SKU の意味が不明 |

## sakura_addon_cdn

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `location` | String | The location of the Addon CDN. | 例は `japaneast` だが、取りうる値の一覧が不明 |
| `patterns` | List of String | The route patterns of the Addon CDN. | `/*` 以外の書式（パス一致、正規表現、ワイルドカードのルール等）が不明 |
| `pricing_level` | Number | The pricing level of the Addon CDN | 取りうる値の一覧、範囲、各レベルの意味が不明 |
| `origin.hostname` | String | The origin host name. | FQDN のみか、IP も可か、ポート指定やプロトコル含むか等の例がない |
| `origin.host_header` | String | The origin host header. | どのような文字列を入れるかの例がない |

## sakura_addon_datalake

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `location` | String | The location of the Addon DataLake | 例は `japaneast` だが、取りうる値の一覧が不明 |
| `performance` | Number | The performance setting of the Addon DataLake. | 単位、取りうる値の一覧、範囲、各値の意味が不明 |
| `redundancy` | Number | The redundancy setting of the Addon DataLake. | 取りうる値の一覧、範囲、各値の意味が不明 |

## sakura_addon_ddos

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `location` | String | The location of the Addon DDoS | 例は `japaneast` だが、取りうる値の一覧が不明 |
| `patterns` | List of String | The route patterns of the Addon DDoS. | `/*` 以外の書式（パス一致、正規表現、ワイルドカードのルール等）が不明 |
| `pricing_level` | Number | The pricing level of the Addon DDoS | 取りうる値の一覧、範囲、各レベルの意味が不明 |
| `origin.hostname` | String | The origin host name. | FQDN のみか、IP も可か、ポート指定やプロトコル含むか等の例がない |
| `origin.host_header` | String | The origin host header. | どのような文字列を入れるかの例がない |

## sakura_addon_dwh

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `location` | String | The location of the Addon DWH | 例は `japaneast` だが、取りうる値の一覧が不明 |

## sakura_addon_etl

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `location` | String | The location of the Addon ETL | 例は `japaneast` だが、取りうる値の一覧が不明 |

## sakura_addon_query

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `location` | String | The location of the Addon Query | 例は `japaneast` だが、取りうる値の一覧が不明 |

## sakura_addon_search

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `location` | String | The location of the Addon Search | 例は `japaneast` だが、取りうる値の一覧が不明 |
| `partition_count` | Number | The partition count of the Addon Search. | 取りうる値の一覧、範囲、step、各値の意味が不明 |
| `replica_count` | Number | The replica count of the Addon Search. | 取りうる値の一覧、範囲、step、各値の意味が不明 |
| `sku` | Number | The SKU of the Addon Search | 取りうる値の一覧、範囲、各 SKU の意味が不明 |

## sakura_addon_streaming

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `location` | String | The location of the Addon Streaming | 例は `japaneast` だが、取りうる値の一覧が不明 |
| `unit_count` | String | The unit count of the Addon Streaming. | なぜ String 型か、取りうる値の一覧/範囲/step、各値の意味（1 ユニットあたりの性能等）が不明 |

## sakura_addon_waf

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `location` | String | The location of the Addon WAF | 例は `japaneast` だが、取りうる値の一覧が不明 |
| `patterns` | List of String | The route patterns of the Addon WAF. | `/*` 以外の書式（パス一致、正規表現、ワイルドカードのルール等）が不明 |
| `pricing_level` | Number | The pricing level of the Addon WAF | 取りうる値の一覧、範囲、各レベルの意味が不明 |
| `origin.hostname` | String | The origin host name. | FQDN のみか、IP も可か、ポート指定やプロトコル含むか等の例がない |
| `origin.host_header` | String | The origin host header. | どのような文字列を入れるかの例がない |

## sakura_object_storage_bucket

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `site_id` | String | The ID of the Object Storage Site. | ID系：site_id の形式（例: `isk01`/`tky01` などのサイトコードか、UUID か）が Description から判断できない。 |

## sakura_object_storage_bucket_cors

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `access_key` | String, Sensitive | The access key for the Object Storage Bucket CORS. | ID/認証情報系：具体的な形式や例がない。 |
| `bucket` | String | The bucket of the Object Storage Bucket CORS. | ID系：バケット名かバケットIDか、Description だけでは判断できない。 |
| `secret_key` | String, Sensitive | The secret key for the Object Storage Bucket CORS. | ID/認証情報系：具体的な形式や例がない。 |
| `cors_rules.allowed_methods` | Set of String | The set of HTTP methods that are allowed to access the origin. | Enum系：取りうる HTTP メソッドの一覧（`GET`, `PUT`, ...）が記載されていない。 |
| `cors_rules.allowed_origins` | Set of String | The set of origins that are allowed to access to the bucket. | URL/ホスト系：入力すべきオリジンの書式例（`https://example.com` など）がない。 |
| `cors_rules.allowed_headers` | Set of String | The set of headers used in `Access-Control-Request-Headers` | 値の例や形式がない。 |
| `cors_rules.expose_headers` | Set of String | The set of headers in the response that users can access from the application. | 値の例や形式がない。 |
| `cors_rules.id` | String | The ID of the CORS rule. | ID系：CORS ルール ID の形式が不明。 |

## sakura_object_storage_bucket_encryption_config

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `kms_key_id` | String | The ID of the KMS key for encryption. | ID系：KMS キー ID の形式（UUID / リソースID 等）が不明。 |
| `site_id` | String | The ID of the Object Storage Site. | ID系：site_id の形式が不明。 |

## sakura_object_storage_bucket_replication_config

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `site_id` | String | The ID of the Object Storage Site. | ID系：site_id の形式が不明。 |

## sakura_object_storage_bucket_versioning

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `access_key` | String, Sensitive | The access key for the Object Storage Bucket Versioning. | ID/認証情報系：具体的な形式や例がない。 |
| `bucket` | String | The bucket of the Object Storage Bucket Versioning. | ID系：バケット名かバケットIDか判断できない。 |
| `secret_key` | String, Sensitive | The secret key for the Object Storage Bucket Versioning. | ID/認証情報系：具体的な形式や例がない。 |

## sakura_object_storage_object

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `access_key` | String, Sensitive | The access key for the Object Storage Object. | ID/認証情報系：具体的な形式や例がない。 |
| `bucket` | String | The bucket of the Object Storage Object. | ID系：バケット名かバケットIDか判断できない。 |
| `key` | String | The key of the Object Storage Object. | Path系：オブジェクトキーの書式例（`path/to/object.txt` 等）がない。 |
| `secret_key` | String, Sensitive | The secret key for the Object Storage Object. | ID/認証情報系：具体的な形式や例がない。 |
| `acl` | String | The ACL of the Object Storage Object. | Enum系：取りうる ACL 値の一覧がない。 |
| `cache_control` | String | The cache control setting for the Object Storage object | 書式例や取りうる値がない。 |
| `content_encoding` | String | The content encoding of the Object Storage Object. | Enum/値系：`gzip` 等の取りうる値の一覧がない。 |
| `content_language` | String | The content language of the Object Storage Object. | Enum/値系：`en`, `ja` 等の取りうる値の一覧がない。 |
| `content_type` | String | The content type of the Object Storage Object. | 書式例（`text/plain`, `application/json` 等）がない。 |
| `source` | String | The path to a file that will be uploaded as the Object Storage Object. | Path系：絶対パス/相対パス、書式例がない。 |

## sakura_object_storage_permission

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `site_id` | String | The ID of the Object Storage Site. | ID系：site_id の形式が不明。 |

## sakura_apigw_cert

問題なし

## sakura_apigw_domain

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `name` | String | Name of the API Gateway Domain | ホスト/URL系：FQDN（例: `api.example.com`）を入力するのか、単なる名称なのかが Description だけでは不明。 |
| `certificate_id` | String | ID of the API Gateway Certificate | ID系：証明書 ID の形式（UUID / 数値 等）が不明。 |

## sakura_apigw_group

問題なし

## sakura_apigw_route

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `protocols` | String | The protocols supported by the Route | Enum系：取りうる値（`http`, `https`）の一覧、および複数指定時の区切り形式（カンマ区切り等）が不明。 |
| `service_id` | String | The Service ID associated with the API Gateway Route | ID系：サービス ID の形式が不明。 |
| `hosts` | List of String | The list of hosts. Auto-issued host or API Gateway Domain can be used. | ホスト系：入力すべきホストの書式例（FQDN）や、Auto-issued host の指定方法が不明。 |
| `https_redirect_status_code` | Number | The HTTPS redirect status code | 数値：取りうるステータスコードの範囲/一覧がない。 |
| `methods` | Set of String | HTTP methods to access the Route | Enum系：取りうる HTTP メソッドの一覧がない。 |
| `path` | String | The path to access the Route. '/' or '~' prefix is required | Path系：`~` プレフィックス時の具体的な書式例や正規表現の指定方法がない。 |
| `regex_priority` | Number | The regex priority | 数値：範囲や step がない。 |
| `groups.id` | String | ID of the API Gateway Group | ID系：グループ ID の形式が不明。 |
| `ip_restriction.ips` | Set of String | The IPv4 addresses to be restricted | IP/CIDR系：CIDR表記（`192.168.0.0/24`）も許容されるのか、単一 IP のみか等が不明。 |
| `ip_restriction.protocols` | String | The protocols to restrict | Enum系：取りうる値の一覧と区切り形式が不明。 |
| `ip_restriction.restricted_by` | String | The category to restrict by | Enum系：取りうる値の一覧（`allowIps` / `denyIps` 等）が不明。 |

## sakura_apigw_service

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `host` | String | The host name of the backend. | ホスト系：入力すべきホストの書式例（FQDN/IP）がない。 |
| `subscription_id` | String | The subscription plan ID associated with the service | ID系：サブスクリプション ID の形式が不明。 |
| `path` | String | The base path for the backend | Path系：先頭に `/` が必要か等の書式例がない。 |
| `cors_config.access_control_allow_headers` | String | Allowed request headers | 書式例/値系：カンマ区切り、`*`、個別ヘッダー名等の指定形式が不明。 |
| `cors_config.access_control_allow_methods` | Set of String | Allowed HTTP methods for CORS | Enum系：取りうる HTTP メソッドの一覧がない。 |
| `cors_config.access_control_allow_origins` | String | Allowed origins for CORS | URL/ホスト系：オリジンの書式例（`https://example.com`, `*` 等）がない。 |
| `cors_config.access_control_exposed_headers` | String | Headers exposed to the client | 書式例/値系：カンマ区切り等の指定形式が不明。 |
| `cors_config.max_age` | Number | Max age for CORS | 数値：単位（秒か）や範囲が不明。 |
| `object_storage_config.endpoint` | String | The object storage endpoint | URL系：エンドポイント URL の書式例がない。 |
| `object_storage_config.region` | String | The object storage region | Enum系：取りうるリージョン値の一覧がない。 |
| `object_storage_config.folder` | String | The folder name within the bucket | Path系：フォルダパスの書式例（先頭/末尾スラッシュの要否等）がない。 |
| `oidc.id` | String | The entity ID of OIDC authentication | ID系：OIDC エンティティ ID の形式が不明。 |

## sakura_apigw_subscription

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `plan_id` | String | Plan ID of the API Gateway Subscription | ID系：プラン ID の形式（UUID / 数値 / プラン名 等）が不明。 |

## sakura_apigw_user

| Field Path | Type | Current Description | 不明瞭な点 |
|---|---|---|---|
| `custom_id` | String | The custom ID of the API Gateway User | ID系：custom_id の形式や制約（文字数、許容文字等）が不明。 |
| `authentication.jwt.algorithm` | String | The JWT algorithm | Enum系：取りうるアルゴリズムの一覧（`HS256`, `RS256` 等）がない。 |
| `authentication.jwt.key` | String | The JWT key | 認証情報系：key の形式や例（kid, 公開鍵ID 等）がない。 |
| `groups.id` | String | ID of the API Gateway Group | ID系：グループ ID の形式が不明。 |
| `ip_restriction.ips` | Set of String | The IPv4 addresses to be restricted | IP/CIDR系：CIDR表記も許容されるか等が不明。 |
| `ip_restriction.protocols` | String | The protocols to restrict | Enum系：取りうる値の一覧と区切り形式が不明。 |
| `ip_restriction.restricted_by` | String | The category to restrict by | Enum系：取りうる値の一覧（`allowIps` / `denyIps` 等）がない。 |
