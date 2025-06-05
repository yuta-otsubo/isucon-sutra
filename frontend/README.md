# ISURIDE Frontend

## Routes

- **/client**
  - ユーザー側が使用するアプリケーションルート
- **/client/history**
  - 履歴
- **/client/contact**
  - 問い合わせ
- **/driver**
  - ドライバー側(配椅子側)が使用するアプリケーションルート


## Setup

```sh
pnpm create remix@latest --template remix-run/remix/templates/spa
```

## Development

```sh
pnpm run dev
```

## Format

```sh
pnpm run fmtcheck # format check
pnpm run fmt # format fix
```

## Production

```sh
pnpm run build
```

## Codegen

```sh
pnpm run codegen
```
