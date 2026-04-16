# Terraform Provider for さくらのクラウド v3

さくらのクラウド向けのTerraform Providerです。

- レジストリ: https://registry.terraform.io/providers/sacloud/sakura
- v2(旧バージョン): https://github.com/sacloud/terraform-provider-sakuracloud

## v3での変更点

変更点は[CHANGES](CHANGES.md)を参照してください。

設定例に関しては[examples](./examples/)にあるtfファイル群を参考にしてください。

## TODO

### v2からの移植

v2からまだ移植できていないリソースのリストになります。これらを利用したい場合にはv2と併用してください。

- archive_share
- cdrom
- certificate_authority
- esme
- mobile_gateway
- sim
- webaccel
- webaccel_activation
- webaccel_acl
- webaccel_certificate

### 新規サービス群の実装

APIGW、IAM、セキュリティコントロール等のリソースの実装。

## 開発者向けドキュメント

開発者向けのドキュメントは[CONTRIBUTING.md](./CONTRIBUTING.md)を参照してください。

