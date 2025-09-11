# webapp/php リソース要約

## 概要

このディレクトリは ISUCON の PHP 実装版の Web アプリケーションです。
Slim Framework を使用したライドシェアアプリケーション（IsuRide）の実装となっています。

## アーキテクチャ

- **フレームワーク**: Slim Framework (PHP)
- **アプリケーション名**: IsuRide
- **データベース**: MySQL
- **ログ**: Monolog
- **コード品質**: PHPStan, PHP_CodeSniffer

## 初回 composer の設定

composer.json があるディレクトリで実施する (webapp/php)
※ composer.json も自動で作られるような気がするが、手動で作成した (composer init で作成できるようだ)

バージョン確認

```sh
composer --version
```

インストールされていることを確認したら、

```sh
composer install
```

```sh
composer update
```

## openapi generator の実行

openapi-generator を使用できるようにインストールする

```sh
https://github.com/openapitools/openapi-generator
```

composer.json があるディレクトリで実施する (webapp/php)

```sh
composer run generate
```

## ディレクトリ構造

```tree
webapp/php/
├── app/                          # アプリケーション設定
│   ├── config.php               # データベース設定、ログ設定
│   ├── middleware.php           # ミドルウェア設定
│   └── routes.php               # ルーティング定義
├── NOTE.md                      # このファイル
├── phpcs.xml                    # PHP_CodeSniffer設定
├── phpstan.neon.dist           # PHPStan設定
├── public/                      # Webサーバー公開ディレクトリ
│   └── index.php               # エントリーポイント
├── src/                        # ソースコード
│   ├── Application/            # アプリケーション層
│   │   ├── Database/          # データベース関連
│   │   │   └── Model/         # データモデル
│   │   │       ├── Chair.php           # 椅子モデル
│   │   │       ├── ChairLocation.php   # 椅子位置モデル
│   │   │       ├── ChairModel.php      # 椅子機種モデル
│   │   │       ├── Owner.php           # オーナーモデル
│   │   │       ├── PaymentToken.php    # 決済トークンモデル
│   │   │       ├── RideRequest.php     # 乗車リクエストモデル
│   │   │       └── User.php            # ユーザーモデル
│   │   └── Payload/           # リクエスト/レスポンス用データ構造
│   │       ├── Coordinate.php          # 座標データ
│   │       └── PostInitializeRequest.php # 初期化リクエスト
│   └── Foundation/            # 基盤層
│       ├── Handlers/          # エラーハンドラー
│       │   ├── HttpErrorHandler.php    # HTTPエラーハンドラー
│       │   └── ShutdownHandler.php     # シャットダウンハンドラー
│       └── ResponseEmitter/   # レスポンス送信
│           └── ResponseEmitter.php     # レスポンス送信処理
└── var/                        # 変数データ
    └── cache/                  # キャッシュディレクトリ
```

## 主要コンポーネント

### 1. エントリーポイント (`public/index.php`)

- Slim Framework アプリケーションの初期化
- ミドルウェアとルーティングの登録
- エラーハンドラーの設定

### 2. 設定ファイル (`app/config.php`)

- データベース接続設定（環境変数対応）
- ログ設定（Monolog 使用）

### 3. データモデル (`src/Application/Database/Model/`)

ライドシェアアプリケーションの主要エンティティ：

- **Chair**: 椅子（車両）情報
- **ChairLocation**: 椅子の位置情報
- **ChairModel**: 椅子の機種情報
- **Owner**: オーナー（運転手）情報
- **User**: ユーザー（乗客）情報
- **RideRequest**: 乗車リクエスト情報
- **PaymentToken**: 決済トークン情報

### 4. 基盤層 (`src/Foundation/`)

- **HttpErrorHandler**: HTTP エラーの統一処理
- **ShutdownHandler**: アプリケーション終了時の処理
- **ResponseEmitter**: レスポンス送信の統一処理

### 5. ペイロード (`src/Application/Payload/`)

- **Coordinate**: 座標データ構造
- **PostInitializeRequest**: 初期化リクエストデータ

## 開発環境設定

- **phpcs.xml**: コーディング規約チェック設定
- **phpstan.neon.dist**: 静的解析設定
- **.gitignore**: キャッシュディレクトリの除外設定

## 環境変数

- `ISUCON_DB_HOST`: データベースホスト（デフォルト: 127.0.0.1）
- `ISUCON_DB_PORT`: データベースポート（デフォルト: 3306）
- `ISUCON_DB_USER`: データベースユーザー（デフォルト: isucon）
- `ISUCON_DB_PASSWORD`: データベースパスワード（デフォルト: isucon）
- `ISUCON_DB_NAME`: データベース名（デフォルト: isuride）

## 特徴

- Clean Architecture の考え方を取り入れた構造
- readonly class を使用したイミュータブルなデータモデル
- 環境変数による設定の外部化
- 統一されたエラーハンドリング
- コード品質管理ツールの導入

## 次のステップ

- ルーティングの実装拡張
- ビジネスロジックの実装
- データベースアクセス層の実装
- API エンドポイントの実装

---

# 181

https://github.com/isucon/isucon14/blob/112fab9cba216969532034d20d2d6e7efe18b618/webapp/php/config.yaml

library: psr-18 に設定が変更されたが、自分の環境では composer run generate に失敗した。
generate コマンドを　`-g php-slim4` から `-g php` に変更することで対応した。
