# Terraform Provider for さくらのクラウド v3

さくら向けTerraform Providerの次期メジャーバージョンとなるv3のリポジトリです。
レジストリ: https://registry.terraform.io/providers/sacloud/sakura


v2: https://github.com/sacloud/terraform-provider-sakuracloud

## v3での変更点

変更点は[CHANGES](CHANGES.md)を参照してください。

## 実装詳細 (開発者向け)

v2からはいくつか実装に関して変更されているところがあります。

### internalディレクトリ

v2では`sakuracloud`ディレクトリにプロバイダーやリソースの実装がフラットに置かれていたが、v3では`internal`以下に移動しています。

- internal/provider: プロバイダ実装
- internal/service: 各ディレクトリにそれぞれのサービスのdata source / resource / model等の実装が置かれている
- internal/common: 各サービスから利用される共通の処理が実装されている。schema / timeout / model等
- internal/validator: 各サービスから利用されるさくら独自のバリデータ群
- internal/test: アクセプタンステストで利用されるヘルパー群

### structure_xxx.goの削減

v2では各リソース毎に`structure_xxx.go`を用意していたが、v3では他と共有される予定のない関数群は各リソースのファイル内に移動しています。
`expandXXX` はresource.go、 `flattenXXX` はmodel.goのように関連の深いファイルに置かれています。

### モデルの実装をmodel.goで共有

v2では`schema.Schema`が全ての共通のインターフェイスになっており実装を共有できたが、Frameworkはそれぞれデータソース・リソース毎にモデルを用意する設計になっているため、処理を共通化しにくい。コピペの実装を防ぐため、data / resourceで共有できる部分は`model.go`に構造体・メソッドを実装し、埋め込みを使って処理を共通化する(主にモデルの更新で使われる)。

### 実装の定義順

実装は以下の順で実装するようになっている

```go
package xxx

import(...)

// リソース向け構造体
type xxxResource {
    client *APIClient // iaas向け。他の独自クライアントを使うサービスの場合は変更する
}

var (
	_ resource.Resource                = &xxxResource{}
	_ resource.ResourceWithConfigure   = &xxxResource{}
	_ resource.ResourceWithImportState = &xxxResource{}
)

// Resourcesで登録するためのヘルパー
func NewXXXResource() resource.Resource {
	return &xxxResource{}
}

func (r *xxxResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_xxx"
}

func (r *xxxResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// clientを設定したり等
}

type xxxResourceModel struct {
	xxxBaseModel  // model.goで実装
	Timeouts timeouts.Value `tfsdk:"timeouts"` // タイムアウトをサポートするには自分で定義に入れる必要がある
}

func (r *xxxResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": common.SchemaResourceId("XXX"),  // SDK v2と違って自分でidを定義する必要がある
            // 他のパラメータ群
			"timeouts": timeouts.Attriutes(ctx, timeouts.Opts{  // タイムアウト向けのパラメータも自分で定義に入れる必要がある
				Create: true, Update: true, Delete: true,
			}),
		},
	}
}

func (r *xxxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *xxxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan xxxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutCreate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

    // Create用の実装

	plan.updateState(xxx)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *xxxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state xxxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read用の実装

	state.updateState(xxx)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *xxxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan xxxxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	//resp.Diagnostics.Append(req.State.Get(ctx, &state)...) // 比較したい場合はstateも使う
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutUpdate(ctx, plan.Timeouts, common.Timeout5min)
	defer cancel()

	// Update用の実装

	plan.updateState(xxx)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *xxxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state xxxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := common.SetupTimeoutDelete(ctx, state.Timeouts, common.Timeout5min)
	defer cancel()

	// Delete用の実装
}

// ヘルパーが必要ならここ以降に書く
```

### ドキュメントの生成

[terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs)を利用。tools以下に利用するためのtools.goを用意してあるので `go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs` を実行した後、以下のコマンド群を入力する。

```
$ tfplugindocs generate --provider-name=sakura
$ ruby tools/update_subcategories.rb
```

これによって `docs` 以下にドキュメントが生成される。 `update_subcategories.rb`によってさくらのマニュアルに合わせたサブカテゴリが各リソースのドキュメントに埋め込まれる。新規リソースを追加した時には `./tools/subcategories.yaml` を更新して適切なサブカテゴリを設定する。
