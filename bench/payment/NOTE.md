## mock_verifier.go の生成

vscode なら `verifier.go` を開いて、`run go generate` を実行することでも作成出来る

```bash
go run go.uber.org/mock/mockgen -typed -source=verifier.go -package=payment -destination=./mock_verifier.go
```
