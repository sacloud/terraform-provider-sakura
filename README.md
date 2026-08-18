# Terraform Provider for さくらのクラウド v3

さくらのクラウドのリソースをTerraformで管理するためのProviderです。
インフラ構成をコード化し、再現性のある運用を行えます。

サポートしているTerraformのバージョンは`1.11`以降です。

## クイックスタート

実践的な構成例は[examples](./examples/)を参照してください。

Terraform 自体については [Terraform の公式ドキュメント](https://developer.hashicorp.com/terraform) を参照してください。
設定方法については [Provider Registry](https://registry.terraform.io/providers/sacloud/sakura) を参照してください。

## v2 からのマイグレーション

> [!IMPORTANT]
> [Terraform Provider for さくらのクラウド v2](https://github.com/sacloud/terraform-provider-sakuracloud) は **2026年12月末をもってメンテナンスを終了**します。
> 詳細は [Terraform Provider v2 メンテナンス終了のお知らせ](https://cloud.sakura.ad.jp/news/2026/06/23/terraform-provider-v2-end-of-maintenance/) をご覧ください。

v2 と *Terraform Provider for さくらのクラウド v3* には互換性がありません。  
v3 における変更点は [CHANGES.md](./CHANGES.md) をご覧ください。

## v3 のリソース対応状況

現在、ほとんどのさくらのクラウドのリソースに対応済みです。

対応していないリソースは以下の通りです。

### v2からの移植

> [!IMPORTANT]
> 以下のリソースは v3 では未移植です。これらを利用している場合は、v3 への移行にあたってそれらのリソースをどう扱うかを検討してください。  
> 例えば、コントロールパネルや API など別の方法で管理する、またはリスクを踏まえたうえで v2 を一時的に利用し続けるなどを検討してください。

- archive_share
- certificate_authority
- esme
- mobile_gateway
- sim

## 開発者向け

本プロジェクトの開発者向けドキュメントは[CONTRIBUTING.md](./CONTRIBUTING.md)を参照してください。

