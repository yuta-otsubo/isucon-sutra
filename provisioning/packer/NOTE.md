### isucon14_base_image.pkr.hcl
ISUCON14用のベースAMI（Amazon Machine Image）をPackerで作成するための設定ファイルです。
Ubuntu 22.04（jammy）をベースに、Ansibleの `application-base.yml` プレイブックで初期セットアップを行い、共通のベースイメージを作成します。

### isucon14.pkr.hcl
ISUCON14の本番用AMIをPackerで作成するための設定ファイルです。
`isucon14_base_image.pkr.hcl` で作成したベースイメージを元に、Ansibleの `application.yml` プレイブックで本番用のセットアップを行い、各チームが利用するAMIを作成します。

### Makefile
PackerによるAMI作成作業を自動化するためのMakefileです。
`make` コマンドで、コミットハッシュの取得、Ansible用ファイルの生成、Packerの初期化・ビルドを一括で実行できます。
