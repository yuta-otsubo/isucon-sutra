
# provisioning/ansible ディレクトリ構造

```
.
├── ansible.cfg
├── application-base.yml
├── application-deploy.yml
├── application.yml
├── benchmark.yml
├── inventory
│   └── localhost
├── make_latest_files.sh
├── NOTE.md
├── README.md
└── roles
    ├── apt
    │   └── tasks
    │       └── main.yml
    ├── base
    │   └── tasks
    │       └── main.yml
    ├── bench
    │   ├── files
    │   └── tasks
    │       └── main.yml
    ├── envcheck
    │   ├── files
    │   │   ├── envcheck.service
    │   │   ├── run-isucon-env-checker.sh
    │   │   └── warmup.sh
    │   └── tasks
    │       └── main.yml
    ├── globalip
    │   ├── files
    │   │   ├── aws-env-isucon-subdomain-address.service
    │   │   └── aws-env-isucon-subdomain-address.sh
    │   └── tasks
    │       └── main.yml
    ├── isuadmin-user
    │   ├── files
    │   │   └── authorized_keys
    │   └── tasks
    │       └── main.yml
    ├── isucon-user
    │   ├── tasks
    │   │   └── main.yml
    │   └── templates
    │       └── env.sh
    ├── mysql
    │   └── tasks
    │       └── main.yml
    ├── nginx
    │   └── files
    │       ├── etc
    │       │   └── nginx
    │       │       ├── sites-available
    │       │       │   └── isuride-php.conf
    │       │       ├── sites-enabled
    │       │       │   └── isuride.conf
    │       │       └── tls
    │       │           ├── _.self.u.isucon.dev.crt
    │       │           ├── _.self.u.isucon.dev.key
    │       │           ├── _.t.isucon.dev.crt
    │       │           ├── _.t.isucon.dev.key
    │       │           ├── _.u.isucon.dev.crt
    │       │           ├── _.u.isucon.dev.issuer.crt
    │       │           └── _.u.isucon.dev.key
    │       └── tasks
    │           └── main.yml
    ├── powerdns
    │   ├── files
    │   │   ├── auth-48
    │   │   ├── pdns.conf
    │   │   ├── pdns.d
    │   │   │   ├── docker.conf
    │   │   │   └── gmysql-host.conf
    │   │   ├── pdns.list
    │   │   └── resolved.conf
    │   └── tasks
    │       └── main.yml
    ├── sandbox.ini
    ├── webapp
    │   ├── files
    │   │   ├── isuride-go.service
    │   │   ├── isuride-node.service
    │   │   ├── isuride-perl.service
    │   │   ├── isuride-php.service
    │   │   ├── isuride-python.service
    │   │   ├── isuride-ruby.service
    │   │   ├── isuride-rust.service
    │   │   └── isuride.php-fpm.conf
    │   └── tasks
    │       ├── go.yaml
    │       ├── main.yml
    │       ├── node.yml
    │       ├── perl.yml
    │       ├── php.yml
    │       ├── python.yml
    │       ├── ruby.yml
    │       └── rust.yml
    ├── xbuild
    │   ├── files
    │   └── tasks
    │       └── main.yml
    └── xbuildwebapp
        ├── files
        │   └── rustup-init.sh
        └── tasks
            └── main.yml

45 directories, 61 files
```

## ファイル・ディレクトリの役割

### 設定ファイル
- **ansible.cfg**: Ansible の基本設定（ロールパス、プロファイル有効化など）
- **inventory/localhost**: ローカル実行用のインベントリファイル（benchmarker, application グループ定義）

### プレイブック
- **application.yml**: アプリケーションサーバの完全セットアップ（全ロール実行）
- **application-base.yml**: ランタイム環境のみセットアップ（Packer ビルド用）
- **application-deploy.yml**: webapp のデプロイのみ実行（既存 webapp 削除→再デプロイ）
- **benchmark.yml**: ベンチマーカーサーバのセットアップ

### スクリプト
- **make_latest_files.sh**: 最新のベンチマーカーバイナリと webapp アーカイブを生成・配置

### ロール（roles/）

#### 基盤系
- **base/**: システム基盤設定（タイムゾーン、SSH、sysctl、pam limits）
- **apt/**: APT パッケージ管理とアップデート

#### ユーザー管理
- **isuadmin-user/**: 管理者ユーザーの作成と SSH キー設定
- **isucon-user/**: ISUCON 競技用ユーザーの作成と環境設定

#### インフラサービス
- **mysql/**: MySQL データベースサーバのセットアップ
- **nginx/**: Nginx ウェブサーバの設定（SSL 証明書、サイト設定含む）
- **powerdns/**: PowerDNS 権威サーバの設定（DNS 解決用）

#### 開発環境
- **xbuild/**: 各種プログラミング言語のビルド環境構築
- **xbuildwebapp/**: webapp 用の追加ビルドツール（Rust など）

#### アプリケーション
- **webapp/**: 各言語版（Go, Node.js, Perl, PHP, Python, Ruby, Rust）の webapp デプロイと systemd サービス設定

#### 運用・監視
- **envcheck/**: 環境チェッカーのセットアップ（競技環境の動作確認用）
- **globalip/**: AWS 環境でのグローバル IP アドレス取得設定
- **bench/**: ベンチマーカーバイナリの配置

#### その他
- **sandbox.ini**: サンドボックス環境用の設定ファイル
