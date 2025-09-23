# NOTE

### bench

このディレクトリで何をやっているのか理解できたら、どういう評価軸でアプリを見ているのかが分かるので、点数アップに繋がる

### Taskfile とは?

- Go 言語で実装されたタスクランナー
- brew install go-task でインストールする必要がある

### TODO

- [ ] 自分の AWS 環境に ECR を作成する: ./Dockerfile

### gen-init-data-sql の実行

- isucon-sutra/Taskfile の `init, up` タスクを実行する
  - コンテナが起動しており、DB に接続できること
  ```
  lsof -i :3306
  ```
  ```
  mysql -h127.0.0.1 -P3306 -uisucon -pisucon -Disuride
  ```
- isucon-sutra/webapp/sql の `init.sh` を実行する</br>初期データが投入される
  ```
  ./init.sh
  ```
- isucon-sutra/bench/ の `gen-init-data-sql` タスクを実行する</br>zip ファイルが作成される
  ```
  task gen-init-data-sql
  ```
- isucon-sutra/webapp/sql の `init.sh` を再度実行する
