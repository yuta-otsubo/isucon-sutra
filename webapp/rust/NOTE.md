# Rust Cargo コマンド説明

## 基本的な Cargo コマンド

### プロジェクトのビルド

```bash
# デバッグビルド
cargo build

# リリースビルド（最適化されたバイナリ）
cargo build --release
```

### プロジェクトの実行

```bash
# デバッグモードで実行
cargo run

# リリースモードで実行
cargo run --release
```

### テストの実行

```bash
# すべてのテストを実行
cargo test

# 特定のテストを実行
cargo test test_name

# テストを並列実行（デフォルト）
cargo test --jobs 4
```

### コードのチェック

```bash
# コンパイルエラーのチェック（バイナリは生成しない）
cargo check

# リリースモードでチェック
cargo check --release
```

### 依存関係の管理

```bash
# 依存関係を追加
cargo add package_name

# 開発依存関係を追加
cargo add --dev package_name

# 依存関係を更新
cargo update

# 依存関係の詳細を表示
cargo tree
```

### ドキュメント

```bash
# ドキュメントを生成
cargo doc

# ドキュメントを生成してブラウザで開く
cargo doc --open
```

### コードフォーマット

```bash
# コードをフォーマット
cargo fmt

# フォーマットをチェック
cargo fmt -- --check
```

### リント

```bash
# コードの品質チェック
cargo clippy

# リントエラーを警告として扱う
cargo clippy -- -D warnings
```

### クリーンアップ

```bash
# ビルド成果物を削除
cargo clean

# 特定のターゲットをクリーン
cargo clean --target x86_64-unknown-linux-gnu
```

## 開発時の便利なコマンド

### ホットリロード（cargo-watch 使用時）

```bash
# ファイル変更を監視して自動ビルド・実行
cargo install cargo-watch
cargo watch -x run

# テストを自動実行
cargo watch -x test
```

### プロファイリング

```bash
# パフォーマンスプロファイリング
cargo install flamegraph
cargo flamegraph

# メモリ使用量の分析
cargo install cargo-expand
cargo expand
```

## Docker 環境での実行

### マルチステージビルド

```dockerfile
# ビルドステージ
FROM rust:1.70 as builder
WORKDIR /usr/src/app
COPY . .
RUN cargo build --release

# 実行ステージ
FROM debian:bullseye-slim
COPY --from=builder /usr/src/app/target/release/app /usr/local/bin/app
CMD ["app"]
```

### コンテナ内での実行

```bash
# Dockerコンテナ内でビルド
docker run --rm -v $(pwd):/app -w /app rust:1.70 cargo build --release

# コンテナ内でテスト実行
docker run --rm -v $(pwd):/app -w /app rust:1.70 cargo test
```

## トラブルシューティング

### よくある問題と解決方法

```bash
# キャッシュをクリアして再ビルド
cargo clean && cargo build

# 依存関係の競合を解決
cargo update

# ターゲットプラットフォームの確認
rustup target list

# ツールチェーンの更新
rustup update
```

### デバッグ情報の表示

```bash
# 詳細なビルド情報を表示
cargo build -vv

# 依存関係の解決過程を表示
cargo tree -v
```
