# Node.js WebApp 構造解析

## ファイル構造

```
/Users/ootsuboyuuta/Work/isucon-sutra/webapp/nodejs/
├── src/
│   ├── types/
│   │   ├── hono.ts          # Honoフレームワーク用の型定義
│   │   └── models.ts        # データベースモデルの型定義
│   ├── app_handlers.ts      # アプリユーザー向けAPIハンドラー
│   ├── chair_handlers.ts    # 椅子（ドライバー）向けAPIハンドラー
│   ├── main.ts             # アプリケーションのエントリーポイント
│   ├── middlewares.ts      # 認証ミドルウェア
│   └── owner_handlers.ts   # オーナー向けAPIハンドラー
├── .gitignore              # Git除外設定
├── biome.json              # Biome設定ファイル
├── package.json            # npm依存関係とスクリプト
├── package-lock.json       # 依存関係のロックファイル
└── tsconfig.js             # TypeScript設定
```

## 各ファイルの役割

### main.ts

- アプリケーションのメインエントリーポイント
- Hono サーバーの初期化とルーティング設定
- MySQL 接続の設定
- 3 つの役割別 API（app/owner/chair）のルート定義

### middlewares.ts

- 認証ミドルウェアの実装
- app_session, owner_session, chair_session の 3 種類の認証
- クッキーベースの認証システム

### types/hono.ts

- Hono フレームワーク用の環境型定義
- データベース接続とユーザー情報の型管理

### types/models.ts

- データベースモデルの型定義
- Chair, User, Owner, Ride 等のエンティティ型

### ハンドラーファイル群

- app_handlers.ts: ユーザー向け API（乗車予約、支払い等）
- chair_handlers.ts: 椅子（ドライバー）向け API（位置情報、乗車状況等）
- owner_handlers.ts: オーナー向け API（売上、椅子管理等）

## 技術スタック解説

### Hono

- 軽量で高速な Web フレームワーク
- Express.js の代替として注目されている
- TypeScript 完全対応
- エッジランタイム対応（Cloudflare Workers 等）
- 特徴：
  - 小さなバンドルサイズ
  - 高いパフォーマンス
  - モダンな API 設計

### Biome

- Rust 製の高速な JavaScript/TypeScript ツールチェーン
- ESLint + Prettier の代替
- 機能：
  - リンティング（コード品質チェック）
  - フォーマッティング（コード整形）
  - インポート整理
- 特徴：
  - 非常に高速な実行速度
  - 設定が簡単
  - 単一ツールで複数機能

## パッケージ管理

### 基本コマンド

```bash
# 依存関係のインストール
npm install

# 開発サーバー起動
npm run dev

# パッケージ追加
npm install <package-name>
npm install -D <package-name>  # 開発依存関係

# パッケージ削除
npm uninstall <package-name>

# 依存関係の更新
npm update

# セキュリティ監査
npm audit
npm audit fix
```

### 主要な依存関係

- **@hono/node-server**: Hono を Node.js 環境で実行するアダプター
- **hono**: Web フレームワーク本体
- **mysql2**: MySQL 接続ライブラリ
- **tsx**: TypeScript 実行環境（開発用）
- **@biomejs/biome**: コード品質ツール
