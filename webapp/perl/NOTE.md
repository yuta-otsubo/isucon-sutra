# perl

## ディレクトリ構成

```
perl/
├── lib/
│   └── Isuride/
│       └── App.pm          # メインアプリケーションモジュール
├── local/                  # 依存モジュールのインストール先
├── .gitignore             # Gitで無視するファイル設定
├── app.psgi               # PSGIアプリケーションエントリーポイント
├── cpanfile               # 依存モジュール定義ファイル
├── cpanfile.snapshot      # 依存モジュールのバージョン固定ファイル
├── Dockerfile             # Dockerコンテナ設定
└── NOTE.md                # 開発メモ
```

### 各ファイルの役割

- **app.psgi**: PSGIアプリケーションのエントリーポイント。Plackサーバーが読み込む
- **cpanfile**: 必要なPerlモジュールを定義。`carton install`で依存関係を解決
- **cpanfile.snapshot**: 依存モジュールのバージョンを固定。再現可能なビルドを保証
- **Dockerfile**: Perlアプリケーション用のコンテナ設定
- **lib/Isuride/App.pm**: Kossyフレームワークを使ったWebアプリケーション本体
- **local/**: `carton install`で依存モジュールがインストールされるディレクトリ


## cpanfile
webapp/perl/cpanfile を参照して、依存モジュールをインストールする。
webapp/perl/ ディレクトリに移動して実行する。

```
carton install
```

※ 何故かすべての依存関係が追加されない。今はまだわからないので後ほど調査する。

## アプリに接続(ローカル)

```
docker build -t isuride-perl .
docker run -p 8080:8080 isuride-perl
```

## .perltidyrc

プロジェクトルートの `.perltidyrc` は Perl コードフォーマッターの設定ファイル。
perltidy コマンドでコードを自動整形する際の規則を定義している。

主な設定内容:
- UTF-8エンコーディング使用
- else文の抱き合わせスタイル
- インデント4スペース
- 最大行長制限なし
- 括弧の密着度設定
- コメントのインデント設定

開発時にコードの統一性を保つために使用される。