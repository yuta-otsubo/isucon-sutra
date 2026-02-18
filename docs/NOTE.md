# docs

## ファイル構成

```
docs/
├── .gitignore                              # Git除外設定
├── .textlintrc.json                        # textlintの設定ファイル
├── .tool-versions                          # asdfツールバージョン管理
├── ISURIDE.md                              # ISURIDEの仕様書
├── manual.md                               # マニュアル
├── NOTE.md                                 # このファイル
├── client-application-simulator-image.png  # クライアントアプリシミュレータの画像
├── package.json                            # Node.js依存関係定義
├── pnpm-lock.yaml                          # pnpmロックファイル
└── prh-rule.yaml                           # prh（校正ツール）のルール定義
```

### 各ファイルの役割

- **ISURIDE.md**: ISURIDEアプリケーションの仕様書
- **manual.md**: 競技マニュアル
- **package.json**: textlint関連の依存関係を管理
- **.textlintrc.json**: Markdown文書の校正ルール設定
- **prh-rule.yaml**: 表記ゆれチェックのルール定義

## pnpm-lockファイルの更新

### package.json更新後のロックファイル更新

```bash
cd docs
pnpm install
```

### 依存関係を最新バージョンに更新

```bash
pnpm update
```

### textlintの実行

```bash
# Lintチェック
pnpm run lint

# 自動修正
pnpm run lint:fix
```
