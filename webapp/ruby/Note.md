# Ruby Webアプリケーション構造

## ファイル構造
```
/Users/ootsuboyuuta/Work/isucon-sutra/webapp/ruby/
|-- lib/
|   `-- isuride/
|       |-- app_handler.rb
|       |-- base_handler.rb
|       |-- chair_handler.rb
|       |-- initialize_handler.rb
|       |-- owner_handler.rb
|       `-- payment_gateway.rb
|-- .dockerignore
|-- .gitignore
|-- config.ru
|-- Dockerfile
|-- Gemfile
|-- Gemfile.lock
`-- Note.md
```

## ファイル要約

### 設定・環境ファイル
- **config.ru**: Rackアプリケーションの設定ファイル。各ハンドラーをマウント
- **Gemfile**: 依存関係定義（Sinatra、MySQL2、Puma等）
- **Gemfile.lock**: 依存関係のロック
- **Dockerfile**: Ruby 3.3.6ベースのコンテナ設定、Pumaサーバー起動
- **.dockerignore/.gitignore**: .bundleディレクトリを除外

### ハンドラー（lib/isuride/）
- **base_handler.rb**: 共通基底クラス。DB接続、エラーハンドリング、料金計算等
- **app_handler.rb**: アプリユーザー向けAPI（ユーザー登録、ライド予約、評価等）
- **chair_handler.rb**: 椅子（車両）向けAPI（椅子登録、位置更新、ライド状態管理）
- **owner_handler.rb**: オーナー向けAPI（オーナー登録、売上確認、椅子管理）
- **initialize_handler.rb**: 初期化API（データベース初期化）
- **payment_gateway.rb**: 決済ゲートウェイとの通信クラス

## 拡張子の違い
- **.rb**: 通常のRubyソースコードファイル（クラス、モジュール定義）
- **.ru**: Rack Up設定ファイル（Rackアプリケーションの起動設定）
## Rackアプリケーションについて
- **Rack**: RubyのWebサーバーとWebアプリケーション間の標準インターフェース
- **config.ru**: Rackアプリケーションのエントリーポイント。URLパスごとに異なるハンドラーをマッピング
- **Sinatra**: RackベースのWebフレームワーク。各ハンドラーはSinatraアプリケーション


## Bundlerコマンド

### Gemfile.lockの更新
```bash
# 依存関係をインストールしてGemfile.lockを更新
bundle install

# 全てのgemを最新バージョンに更新してGemfile.lockを再生成
bundle update
```

- `bundle install`: Gemfileに記載された依存関係をインストール。Gemfile.lockが存在する場合はそのバージョンを使用
- `bundle update`: 全てのgemを最新バージョンに更新し、Gemfile.lockを完全に再生成
