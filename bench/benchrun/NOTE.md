# code genarate

Bufツールを使用してGoのコードを自動生成するようだ

## ファイルの作成

benchrun/buf.yaml
benchrun/buf.gen.yaml
benchrun/env.go
benchrun/generate.go
benchrun/proto/isuxportal/*

proto/isuxportal/* 配下のファイルを元に、benchrun/gen/isuxportal/* 配下にGoのコードを生成する

## コード生成

```
brew install bufbuild/buf/buf
```

generate.go を実行する

## .proto ファイルとは？

**概要説明**
`.proto` ファイルは、Protocol Buffers（プロトコルバッファー、略して「protobuf」）というデータシリアライズ（構造化データの保存や転送方式）用の定義ファイルです。Googleが開発したもので、主に異なる言語間やサービス間で効率よくデータ構造をやりとりするための「設計図」のような役割を果たします。

**用途と特徴**
`.proto` ファイルには、データの構造（たとえばメッセージやサービスなど）を宣言的に記述します。このファイルをもとに、さまざまなプログラミング言語向けのコード（クラスや構造体など）を自動生成できるため、API通信やマイクロサービス間連携、gRPCなどでよく利用されます。

**まとめ**
要するに、`.proto` ファイルは「サービスやアプリ間でやりとりするデータの形やルールを決めるファイル」です。これがあることで、異なる環境同士でもデータのやりとりを効率かつ安全に行えます。
