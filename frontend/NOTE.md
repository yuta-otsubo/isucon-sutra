# Frontend ディレクトリ構成説明

## 設定ファイル
- `package.json`: プロジェクトの依存関係と実行スクリプトを定義
- `package-lock.json`: 依存関係の正確なバージョンを固定するためのファイル
- `pnpm-lock.yaml`: pnpmパッケージマネージャ用の依存関係ロックファイル
- `tsconfig.json`: TypeScriptの設定ファイル
- `vite.config.ts`: Viteビルドツールの設定ファイル
- `tailwind.config.ts`: Tailwind CSSフレームワークの設定ファイル
- `postcss.config.js`: PostCSSの設定ファイル
- `openapi-codegen.config.ts`: OpenAPI仕様からTypeScriptコードを生成するための設定
- `.eslintrc.cjs`: ESLintコード静的解析ツールの設定ファイル
- `.prettierrc.js`: Prettierコードフォーマッタの設定ファイル
- `.prettierignore`: Prettierが無視すべきファイルを指定
- `.gitignore`: Gitが無視すべきファイルを指定
- `.DS_Store`: macOSが自動生成するシステムファイル（バージョン管理対象外）

## その他
- `README.md`: フロントエンドプロジェクトの説明と使用方法を記載したドキュメント

---

- `vite`とは:
    - フロントエンドのビルドツール
    - モダンなフロントエンド開発をサポートするツール
    - 高速な開発体験を提供
    - バンドル、圧縮、コード分割などの機能を提供
    - ホットリローディング、TypeScriptサポート、ESMサポートなどの機能を提供
    - バンドル、圧縮、コード分割などの機能を提供
- `ホットリローディング`とは:
    - コードの変更を即座にブラウザに反映する
    - ファイルの変更を検知して自動的にブラウザを更新する

- `tailwind`とは:
    - ユーティリティクラスベースのCSSフレームワーク
    - クラスを追加することでスタイルを適用できる

---

## 詳細ディレクトリ構造

```
frontend/
│
├── app/                         # アプリケーションのメインコード
│   ├── apiClient/               # API通信用のクライアントコード
│   │   ├── apiFetcher.ts        # APIリクエスト送信の基本機能
│   │   ├── apiComponents.ts     # APIコンポーネント定義
│   │   ├── apiParameters.ts     # APIパラメータ型定義
│   │   ├── apiSchemas.ts        # APIレスポンスのスキーマ定義
│   │   └── apiContext.ts        # APIコンテキスト管理
│   │
│   ├── components/              # 再利用可能なUIコンポーネント
│   │   ├── primitives/          # 基本的なUIコンポーネント
│   │   │   ├── avatar/          # ユーザーアバター関連コンポーネント
│   │   │   │   └── avatar.tsx   # アバターコンポーネント
│   │   │   │
│   │   │   ├── button/          # ボタン関連コンポーネント
│   │   │   │   └── button.tsx   # ボタンコンポーネント
│   │   │   │
│   │   │   ├── modal/           # モーダル関連コンポーネント
│   │   │   │   └── modal.tsx    # モーダルコンポーネント
│   │   │   │
│   │   │   └── rating/          # 評価関連コンポーネント
│   │   │       └── rating.tsx   # 評価コンポーネント
│   │   │
│   │   ├── hooks/               # カスタムReactフック
│   │   │   └── useOnClickOutside.ts  # 要素外クリック検出フック
│   │   │
│   │   ├── icon/                # アイコンコンポーネント
│   │   │   ├── circle.tsx       # 円形アイコンコンポーネント
│   │   │   └── type.ts          # アイコン型定義
│   │   │
│   │   └── FooterNavigation.tsx # フッターナビゲーション
│   │
│   ├── routes/                  # アプリケーションのルート定義
│   │   ├── _index/              # トップページルート
│   │   │   └── route.tsx        # トップページコンポーネント
│   │   │
│   │   ├── client/              # クライアントルート
│   │   │   ├── route.tsx        # クライアントページコンポーネント
│   │   │   └── userProvider.tsx # ユーザー情報プロバイダー
│   │   │
│   │   ├── client._index/       # クライアントインデックスページ
│   │   │   └── route.tsx        # クライアントインデックスコンポーネント
│   │   │
│   │   ├── client.account/      # クライアントアカウントページ
│   │   │   └── route.tsx        # クライアントアカウントコンポーネント
│   │   │
│   │   ├── client.history/      # クライアント履歴ページ
│   │   │   └── route.tsx        # クライアント履歴コンポーネント
│   │   │
│   │   ├── client_contact/      # クライアント連絡先ページ
│   │   │   └── route.tsx        # クライアント連絡先コンポーネント
│   │   │
│   │   ├── driver/              # ドライバールート
│   │   │   ├── route.tsx        # ドライバーページコンポーネント
│   │   │   └── driverProvider.tsx # ドライバー情報プロバイダー
│   │   │
│   │   ├── driver._index/       # ドライバーインデックスページ
│   │   │   └── route.tsx        # ドライバーインデックスコンポーネント
│   │   │
│   │   └── driver.history/      # ドライバー履歴ページ
│   │       └── route.tsx        # ドライバー履歴コンポーネント
│   │
│   ├── entry.client.tsx         # クライアントエントリーポイント
│   ├── root.tsx                 # アプリケーションのルートコンポーネント
│   └── tailwind.css             # Tailwind CSSスタイル定義
│
├── node_modules/                # NPM依存関係
│
├── public/                      # 静的ファイル
│   ├── favicon.ico              # サイトのファビコン
│   ├── logo-dark.png            # ダークモードロゴ
│   └── logo-light.png           # ライトモードロゴ
│
└── [その他設定ファイル]         # 上記の「設定ファイル」セクションで説明されている項目
```

### 主要なディレクトリとファイルの役割

1. **app/apiClient/**
   - API通信のためのコード。OpenAPI仕様から自動生成されたクライアント
   - TypeScriptの型安全性を保ちながらバックエンドと通信するための機能を提供

2. **app/components/**
   - 再利用可能なUIコンポーネント。アプリケーション全体で使われる
   - primitives/: ボタンやモーダルなどの基本UI要素
   - hooks/: 共通のReactロジックを抽出したカスタムフック
   - icon/: アイコンコンポーネント

3. **app/routes/**
   - アプリケーションのページとルーティング
   - ファイル名がURLパスと対応する規則に基づいたルーティング構造
   - client/: クライアント(利用者)向けのページ
   - driver/: ドライバー(運転手)向けのページ
   - ネストされたルートは親子関係を表現（例: client.account/）

4. **public/**
   - ブラウザから直接アクセス可能な静的リソース
   - アプリケーションのロゴや favicon などのアセットを格納
