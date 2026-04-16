# Terraform Provider for さくらのクラウド v3

さくらのクラウドのリソースをTerraformで管理するためのProviderです。
インフラ構成をコード化し、再現性のある運用を行えます。

## クイックスタート

```terraform
terraform {
  required_providers {
    sakura = {
      source = "sacloud/sakura"

      # We recommend pinning to the specific version of the Sakura Provider you're using
      # since new versions are released frequently
      version = "3.0.0"
      #version = "~> 3"
    }
  }
}

# Configure the Sakura Provider
provider "sakura" {
  # More information on the authentication methods supported by
  # the Sakura Provider can be found here:
  # https://docs.usacloud.jp/terraform/provider/

  # profile = "..."
}
```

より実践的な構成例は[examples](./examples/)を参照してください。

Terraform 自体については [Terraform の公式ドキュメント](https://developer.hashicorp.com/terraform) を参照してください。
設定方法については [Provider Registry](https://registry.terraform.io/providers/sacloud/sakura) を参照してください。

## v2 からのマイグレーション

[Terraform Provider for さくらのクラウド v2](https://github.com/sacloud/terraform-provider-sakuracloud) と *Terraform Provider for さくらのクラウド v3* には
互換性がありません。

v3 における変更点は [CHANGES.md](./CHANGES.md) をご覧ください。

## v3 のリソース対応状況

現在、ほとんどのさくらのクラウドのリソースに対応済みです。

対応していないリソースは以下の通りです。

### v2からの移植

以下のリソースは未移植です。必要な場合はv2との併用を検討してください。

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

APIGW、IAM、セキュリティコントロール等の最近リリースされたリソースは一部実装されていません。

## 開発者向け

本プロジェクトの開発者向けドキュメントは[CONTRIBUTING.md](./CONTRIBUTING.md)を参照してください。

