# 3-initial-data.sql.gz の作成手順

## 手順

1. Docker Composeでdbコンテナを起動:
```bash
cd /Users/ootsuboyuuta/Work/isucon-sutra/development
docker compose -f compose-local.yml ps  # 起動確認
docker compose -f compose-local.yml up -d db  # 起動していない場合
```

2. asdfでMySQLクライアントをインストール:
```bash
asdf plugin add mysql
asdf install mysql 8.0.34
asdf set -p mysql 8.0.34
```

2. スキーマとマスターデータをロードした状態でダンプを作成:
```bash
cd /Users/ootsuboyuuta/Work/isucon-sutra/webapp/sql
mysqldump -h127.0.0.1 -P3306 -uroot -pisucon isuride | gzip > 3-initial-data.sql.gz
```

2. init.shで全データをロード（確認用）:
```bash
ISUCON_DB_HOST=127.0.0.1 ISUCON_DB_USER=root ISUCON_DB_PASSWORD=isucon ISUCON_DB_NAME=isuride ./init.sh
```

## 注意
- `3-initial-data.sql.gz`が存在しない状態で`init.sh`を実行すると、最後のステップでエラーになります
- 先にダンプを作成してから`init.sh`で全体の動作を確認してください
