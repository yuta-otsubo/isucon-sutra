# Python WebApp ディレクトリ構造

## ファイル構造

```
webapp/python/
├── .venv/                    # 仮想環境ディレクトリ
├── app/                      # メインアプリケーションパッケージ
│   ├── __init__.py          # パッケージ初期化ファイル
│   ├── main.py              # FastAPIアプリケーションのメインファイル
│   ├── models.py            # データモデル定義
│   ├── app_handlers.py      # アプリユーザー向けAPIハンドラー
│   ├── chair_handlers.py    # 椅子向けAPIハンドラー
│   ├── owner_handlers.py    # オーナー向けAPIハンドラー
│   ├── payment_gateway.py   # 決済ゲートウェイ処理
│   ├── middlewares.py       # ミドルウェア定義
│   ├── utils.py             # ユーティリティ関数
│   └── sql.py               # SQL関連処理
├── build/                   # ビルド成果物ディレクトリ
├── isuride.egg-info/        # パッケージ情報ディレクトリ
├── .mypy_cache/             # mypyキャッシュディレクトリ
├── .nox/                    # noxキャッシュディレクトリ
├── .ruff_cache/             # ruffキャッシュディレクトリ
├── __pycache__/             # Pythonキャッシュディレクトリ
├── .gitignore               # Git除外設定ファイル
├── .python-version          # Pythonバージョン指定ファイル (3.13)
├── .pre-commit-config.yaml  # pre-commit設定ファイル
├── NOTE.md                  # このファイル
├── README.md                # 開発環境構築手順
├── pyproject.toml           # プロジェクト設定と依存関係定義
├── noxfile.py               # noxタスク定義ファイル
├── ruff.toml                # ruff設定ファイル
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

- **noxfile.py**: nox タスクランナーの設定ファイル

  - uv をバックエンドとして使用
  - コード品質チェック（lint）セッション
  - 型チェック（mypy）セッション
  - Python 3.10 環境での自動実行

### アプリケーションコード

- **app/main.py**: FastAPI アプリケーションのメインファイル

  - データベース接続設定
  - API エンドポイント定義
  - アプリケーション初期化処理

- **app/models.py**: データモデル定義

  - Pydantic モデル定義
  - データベーステーブル構造に対応したモデル

- **app/app_handlers.py**: アプリユーザー向けAPIハンドラー

  - ユーザー登録・認証
  - 乗車リクエスト・評価
  - 決済処理
  - 通知機能

- **app/chair_handlers.py**: 椅子向けAPIハンドラー

  - 椅子登録・認証
  - 位置情報更新
  - 乗車リクエスト受付・完了

- **app/owner_handlers.py**: オーナー向けAPIハンドラー

  - オーナー登録・認証
  - 椅子管理
  - 売上確認

- **app/payment_gateway.py**: 決済ゲートウェイ処理

  - 決済API連携
  - 決済状態管理

- **app/middlewares.py**: ミドルウェア定義

  - 認証ミドルウェア
  - ログ出力ミドルウェア

- **app/utils.py**: ユーティリティ関数

  - 共通処理関数
  - ヘルパー関数

- **app/sql.py**: SQL関連処理

  - データベース接続
  - SQL実行関数

- **app/__init__.py**: Python パッケージとして認識させるための空ファイル

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

## ruff について

### ruff とは

ruff は、Rust で書かれた高速な Python リンター・フォーマッターです。従来の flake8、isort、black などの複数ツールの機能を統合し、大幅な高速化を実現しています。

### 主な特徴

- **高速性**: Rust で実装されているため、従来の Python ツールより 10-100 倍高速
- **統合性**: リント、フォーマット、インポート整理を一つのツールで実行
- **互換性**: flake8、isort、black などの既存ルールと互換性
- **設定の簡単さ**: pyproject.toml または ruff.toml で設定可能

### 主要な ruff コマンド

```bash
# コードのリント（問題点の検出）
ruff check .

# コードのフォーマット
ruff format .

# リントとフォーマットを同時実行
ruff check --fix . && ruff format .

# 特定のルールのみチェック
ruff check --select E,W .

# 特定のルールを無視
ruff check --ignore E501 .
```

### 設定ファイル

**ruff.toml** では以下の設定が可能：

```toml
# 行の最大長
line-length = 88

# チェックするルール
select = ["E", "W", "F", "I"]

# 無視するルール
ignore = ["E501"]

# 除外するファイル・ディレクトリ
exclude = [".venv", "__pycache__"]
```

### 従来のツールとの比較

| 機能           | ruff       | flake8 + black + isort |
| -------------- | ---------- | ---------------------- |
| 速度           | 非常に高速 | 低速                   |
| 設定           | 統一       | 個別設定が必要         |
| インストール   | 単一       | 複数パッケージ         |
| メンテナンス   | 簡単       | 複雑                   |

ruff は特に大規模なプロジェクトや CI/CD 環境で、コード品質チェックの実行時間を大幅に短縮できる利点があります。

## nox タスクランナーについて

### nox とは

nox は、Python プロジェクトのタスク自動化ツールです。Makefile や tox の代替として使用され、以下の特徴があります：

- **複数 Python バージョン対応**: 異なる Python バージョンでのテスト実行
- **仮想環境自動管理**: 各セッションで独立した仮想環境を作成
- **タスク定義**: Python コードでタスクを定義し、複雑なワークフローを管理
- **依存関係管理**: セッションごとに必要なパッケージを自動インストール
- **CI/CD 統合**: GitHub Actions などの CI/CD 環境での利用に最適

### 主要な nox コマンド

```bash
# 利用可能なセッション一覧を表示
nox --list

# 特定のセッションを実行
nox --session lint
nox --session mypy

# 全セッションを実行
nox

# セッションを再実行（キャッシュを無視）
nox --reuse-existing-env

# 特定のPythonバージョンで実行
nox --python 3.10
```

### 現在のプロジェクトでの nox 設定

**noxfile.py** では uv をバックエンドとして使用し、以下のセッションが定義されています：

```python
nox.options.default_venv_backend = "uv"
```

1. **lint セッション**:

   - uv を使用して pre-commit をインストール
   - 全ファイルに対して linting を実行
   - Python 3.10 環境で実行
   - `uv run` でコマンド実行

2. **mypy セッション**:
   - uv sync で依存関係を同期
   - uv を使用して mypy をインストール
   - isuride パッケージ全体の型安全性を検証
   - Python 3.10 環境で実行
   - `uv run` でコマンド実行

### 開発ワークフローでの活用

```bash
# コード品質チェック
nox --session lint

# 型チェック
nox --session mypy

# 全チェック実行
nox
```

### 従来のツールとの比較

| 機能                  | nox        | Makefile | tox        | 手動実行 |
| --------------------- | ---------- | -------- | ---------- | -------- |
| Python バージョン管理 | 自動       | 手動     | 自動       | 手動     |
| 仮想環境管理          | 自動       | 手動     | 自動       | 手動     |
| 依存関係管理          | 自動       | 手動     | 自動       | 手動     |
| タスク定義            | Python     | Shell    | INI        | 手動     |
| CI/CD 統合            | 優れている | 基本     | 優れている | 困難     |

nox は特に複数の Python バージョンでのテストや、複雑な開発ワークフローの管理に適しています。

### nox と uv の統合

本プロジェクトでは nox と uv を組み合わせて使用しています：

- **`nox.options.default_venv_backend = "uv"`**: nox が uv をバックエンドとして使用
- **`session.run("uv", ..., external=True)`**: uv コマンドを直接実行
- **高速化**: uv の高速なパッケージインストールを活用
- **一貫性**: プロジェクト全体で uv を統一使用

```bash
# nox セッションでの uv 使用例
# 全件実行
uv run --link-mode=copy nox
```
