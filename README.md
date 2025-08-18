# Terraform Provider for SakuraCloud v3

さくら向けTerraform Providerの次期メジャーバージョンとなるv3の開発リポジトリです

## v3での変更点

### terraform-plugin-frameworkによる書き換え

現在リリースされているv2はSDK v2を利用していますが、v3はFrameworkを利用しています。
muxなども使っておらず、完全移行となります。

### いくつかの命名の変更

- `sakuracloud_`プレフィックスは`sakura_`となります。環境変数などを設定することで過去のプレフィックスもサポートする予定です。
- 必要なものはリソース名も適切なものに変更される可能性があります。

### 使われてない機能の削除

- dataの`filter`等、サービスによってはあまりに使われてないものは削除される予定です。

### 未決定のもの

- terraformが記述を属性に統一しようとしているが、ブロックの記述をいつまでサポートするべきか。`timeout`等は属性ベースにすると設定の更新が必要。

```hcl
# 属性ベースの書き方。こちらが推奨されている
map = {
    //...
}

# ブロックでの書き方。こちらは古い書き方で、現状Blockを使うことで実装できる
map {
    // ...
}
```

## 実装詳細 (開発者向け)

v2からはいくつか実装に関して変更されているところがあります。

### internal/providerディレクトリ

v2では`sakuracloud`ディレクトリにプロバイダーの実装が置かれていたが、`internal/provider`に移動しています。

### structure_xxx.goの削減

v2では各リソース毎に`structure_xxx.go`を用意していたが、v3では他と共有される予定のない関数群は各リソースのファイル内に移動しています。
主に `expandXXX` や `flattenXXX` のような関数群が対象となっています。

### モデルの実装をmodel.goで共有

v2では`schema.Schema`が全ての共通のインターフェイスになっており実装を共有できたが、Frameworkはそれぞれリソース毎にモデルを用意する設計になっているため、処理を共通化しにくい。コピペの実装を防ぐため、data / resourceで共有できる部分は`model.go`に構造体・メソッドを実装し、埋め込みを使って処理を共通化する(主にモデルの更新で使われる)。
モデルの数が増えてきたら`model.go`をディレクトリ以下に移して分割することも考慮する。

### 実装の定義順

実装は以下の順で実装するようになっている

```go
package sakura

import(...)

// リソース向け構造体
type xxxResource {
    client *APIClient // iaas向け。他の独自クライアントを使うサービスの場合は変更する
}

var (
	_ resource.Resource                = &xxxResource{}
	_ resource.ResourceWithConfigure   = &xxxResource{}
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
	sakuraXXXBaseModel  // model.goで実装
	Timeouts timeouts.Value `tfsdk:"timeouts"` // タイムアウトをサポートするには自分で定義に入れる必要がある
}

func (r *xxxResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schemaResourceId("XXX"),  // SDK v2と違って自分でidを定義する必要がある
            // 他のパラメータ
		},
		Blocks: map[string]schema.Block{  // タイムアウト向けのパラメータも自分で定義に入れる必要がある
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
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

    // Create用の実装

    // Readを呼び出して状態を更新する用のヘルパー
	updateResourceByRead(ctx, r, &resp.State, &resp.Diagnostics, xxx.ID.String())
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
	if resp.Diagnostics.HasError() {
		return
	}

	// Update用の実装

	updateResourceByRead(ctx, r, &resp.State, &resp.Diagnostics, key.ID.String())
}

func (r *xxxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state xxxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete用の実装

	resp.State.RemoveResource(ctx)　　// SDK v2ではd.SetId("")に相当
}

// ヘルパーが必要ならここ以降に書く
```

そのほか主要なファイルの説明は以下

- provider.go: プロバイダーのそのものの実装。DataSources/Resourcesに各リソースを登録する。
- strcuture.go: data / resourceでよく使われるデータ変換向けヘルパー群
- data_source_schema.go/resource_schema.go: 各リソースで共通でよく使われるスキーマの定義群
- validators.go: パラメータのバリデーションで使う独自バリデータ群