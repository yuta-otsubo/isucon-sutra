# Python WebApp ディレクトリ構造

## ファイル構造

```
webapp/python/
├── .venv/                    # 仮想環境ディレクトリ
├── src/
│   └── isuride/             # メインアプリケーションパッケージ
│       ├── __init__.py      # パッケージ初期化ファイル
│       ├── app.py           # FastAPIアプリケーションのメインファイル
│       └── py.typed         # 型ヒントサポートファイル
├── .gitignore               # Git除外設定ファイル
├── .python-version          # Pythonバージョン指定ファイル (3.10)
├── NOTE.md                  # このファイル
├── README.md                # 開発環境構築手順
├── pyproject.toml           # プロジェクト設定と依存関係定義
└── uv.lock                  # uv依存関係ロックファイル
```

## 各ファイルの役割

### 設定ファイル

- **pyproject.toml**: プロジェクトのメタデータ、依存関係、ビルド設定を定義

  - プロジェクト名: `isuride`
  - Python 3.10 以上が必要
  - 主要依存関係: FastAPI、SQLAlchemy、PyMySQL、cryptography、python-ulid
  - 開発依存関係: gunicorn

- **uv.lock**: uv パッケージマネージャーが生成する依存関係のロックファイル

  - 正確なバージョンとハッシュ値を記録
  - 再現可能なビルドを保証

- **.python-version**: Python バージョン 3.10 を指定

### アプリケーションコード

- **src/isuride/app.py**: FastAPI アプリケーションのメインファイル

  - データベース接続設定
  - API エンドポイント定義
  - オーナー、チェア、アプリユーザーの登録・管理機能
  - 支払い、通知、座標管理などの機能

- **src/isuride/**init**.py**: Python パッケージとして認識させるための空ファイル

- **src/isuride/py.typed**: 型ヒントサポートを有効にするための空ファイル

### 開発環境

- **.venv/**: uv が作成する仮想環境ディレクトリ
- **README.md**: 開発環境構築とアプリケーション実行手順

## uv パッケージマネージャーについて

### uv とは

uv は、Rust で書かれた高速な Python パッケージマネージャーです。pip、pipenv、poetry の代替として設計されており、以下の特徴があります：

- **高速性**: Rust で実装されているため、従来の Python パッケージマネージャーより大幅に高速
- **依存関係解決**: 効率的な依存関係解決アルゴリズム
- **ロックファイル**: 再現可能なビルドのためのロックファイル生成
- **仮想環境管理**: 自動的な仮想環境作成と管理
- **プロジェクト管理**: pyproject.toml ベースのプロジェクト設定

### 主要な uv コマンド

```bash
# 依存関係のインストール
uv sync

# 仮想環境の作成と依存関係のインストール
uv venv

# パッケージの追加
uv add <package-name>

# 開発依存関係の追加
uv add --dev <package-name>

# アプリケーションの実行
uv run <command>
uv run python src/isuride/app.py

# 開発サーバーの起動
uv run fastapi dev src/isuride/app.py

# 依存関係の更新
uv lock --upgrade

# 仮想環境の削除
uv venv remove

# パッケージの削除
uv remove <package-name>
```

### 開発ワークフロー

1. **環境構築**:

   ```bash
   uv sync
   ```

2. **開発サーバー起動**:

   ```bash
   uv run fastapi dev src/isuride/app.py
   ```

3. **ベンチマーク実行**:
   ```bash
   go run . run --target http://localhost:8000 -t 60
   ```

### 従来のツールとの比較

| 機能           | uv         | pip      | pipenv | poetry   |
| -------------- | ---------- | -------- | ------ | -------- |
| 速度           | 非常に高速 | 低速     | 中速   | 中速     |
| 依存関係解決   | 高速       | 基本     | 中速   | 中速     |
| ロックファイル | あり       | なし     | あり   | あり     |
| 仮想環境管理   | 自動       | 手動     | 自動   | 自動     |
| pyproject.toml | 完全対応   | 部分対応 | 非対応 | 完全対応 |

uv は特に大規模なプロジェクトや CI/CD 環境で、依存関係のインストール時間を大幅に短縮できる利点があります。
