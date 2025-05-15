# Frontend ディレクトリ構成説明

## ディレクトリ
- `app/`: アプリケーションのメインコード。コンポーネント、ルート、APIクライアントなどを含む
- `node_modules/`: npmパッケージの依存関係が格納されるディレクトリ
- `public/`: 静的ファイル（画像、フォントなど）を格納するディレクトリ

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
