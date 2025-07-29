# 自己証明書の作成

ドメインに合わせて作成

```bash
openssl req -new -newkey rsa:2048 -days 365 -nodes -x509 \
  -subj "/CN=*.isuconsutra.dev" \
  -keyout _.self.u.isuconsutra.dev.key \
  -out _.self.u.isuconsutra.dev.crt
```

パブリックキーなので、別途作成して SSM パラメータストアに暗号ファイルとして配置する
