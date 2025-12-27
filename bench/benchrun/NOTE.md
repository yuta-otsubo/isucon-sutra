# code genarate

Bufツールを使用してGoのコードを自動生成するようだ

## ファイルの作成

benchrun/buf.yaml
benchrun/buf.gen.yaml
benchrun/env.go
benchrun/generate.go
benchrun/proto/isuxportal/*

/benchrun/proto/isuxportal/*.proto ファイルを元に、/benchrun/gen/isuxportal/* 配下にGoのコードを生成する

※isuxportal/*.proto ファイルは自分で用意する必要がある

## コード生成

```
brew install bufbuild/buf/buf
```

generate.go を実行する

## .proto ファイルとは？

**概要説明**
`.proto` ファイルは、Protocol Buffers（プロトコルバッファー、略して「protobuf」）というデータシリアライズ（構造化データの保存や転送方式）用の定義ファイルです。Googleが開発したもので、主に異なる言語間やサービス間で効率よくデータ構造をやりとりするための「設計図」のような役割を果たします。

**用途と特徴**
`.proto` ファイルには、データの構造（たとえばメッセージやサービスなど）を宣言的に記述します。このファイルをもとに、さまざまなプログラミング言語向けのコード（クラスや構造体など）を自動生成できるため、API通信やマイクロサービス間連携、gRPCなどでよく利用されます。

**まとめ**
要するに、`.proto` ファイルは「サービスやアプリ間でやりとりするデータの形やルールを決めるファイル」です。これがあることで、異なる環境同士でもデータのやりとりを効率かつ安全に行えます。


## proto配下のファイルの役割説明

proto配下のファイルは、ISUCONポータルシステムのgRPC API定義を構成しています。以下のような構造で役割分担されています：

### 📁 **ルートレベル**
- **`error.proto`** - 共通エラーメッセージの定義（エラーコード、メッセージ、デバッグ情報）

### 📁 **resources/ (リソース定義)**
データモデル（エンティティ）の定義ファイル群：

- **`team.proto`** - チーム情報（ID、名前、メンバー、学生ステータスなど）
- **`contestant.proto`** - 参加者情報
- **`benchmark_job.proto`** - ベンチマークジョブの状態管理（PENDING/RUNNING/FINISHED等）
- **`benchmark_result.proto`** - ベンチマーク実行結果
- **`contest.proto`** - コンテスト情報
- **`leaderboard.proto`** - リーダーボード情報
- **`contestant_instance.proto`** - 参加者のインスタンス情報
- **`clarification.proto`** - 質問・回答
- **`notification.proto`** - 通知システム
- **`env_check.proto`** - 環境チェック結果
- その他（`coupon.proto`, `staff.proto`, `survey_response.proto`）

### 📁 **services/ (API サービス定義)**
各機能ごとのgRPCサービス定義：

#### **bench/** - ベンチマーク実行系
- **`reporting.proto`** - ベンチマーク結果の報告API
- **`receiving.proto`** - ベンチマークジョブの受信API
- **`cancellation.proto`** - ベンチマークキャンセルAPI

#### **admin/** - 管理者機能
- **`benchmark.proto`** - ベンチマーク管理
- **`teams.proto`** - チーム管理
- **`clarifications.proto`** - 質問管理
- **`dashboard.proto`** - 管理ダッシュボード
- **`leaderboard_dump.proto`** - リーダーボードデータエクスポート
- その他管理機能

#### **contestant/** - 参加者機能
- **`benchmark.proto`** - ベンチマーク実行
- **`dashboard.proto`** - 参加者ダッシュボード
- **`clarifications.proto`** - 質問機能
- **`notification.proto`** - 通知受信
- その他参加者機能

#### **common/** - 共通機能
- **`me.proto`** - 自分の情報取得
- **`storage.proto`** - ファイルストレージ

#### **audience/** & **registration/** - その他機能
- 観戦者機能と登録機能

### 📁 **misc/** - その他
- **`leaderboard_etag.proto`** - リーダーボードのキャッシュ制御
- **`bypass_token.proto`** - 認証バイパストークン
- **`bot/`** - ボット関連定義

## まとめ

この構造は**ドメイン駆動設計**の考え方に基づいており：
- **resources** = データモデル（Entity/Value Object）
- **services** = アプリケーションサービス（Use Case）
- 役割ごとに明確に分離されており、保守性と可読性が高い設計になっています

# frontend_hashes.json の生成

`frontend_hashes.json` は、フロントエンドのビルド時にViteプラグインによって自動生成される。

## 生成コマンド

```bash
cd frontend
make build
```

または直接：

```bash
cd frontend
pnpm run build
```

## 用途

ビルドされたフロントエンドファイルのMD5ハッシュ値を記録し、ベンチマーク時にファイルの整合性チェックに使用される。
