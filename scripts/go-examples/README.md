# 可重复的 Go 正文验证

从书仓根目录运行：

```bash
go run scripts/verify-go-examples.go part-1-foundations/payment-lifecycle.md part-1-foundations/routing-system-boundaries.md
```

脚本按 `manifest.json` 从当前 Markdown 提取每个 Go 代码块，应用最小编译包装与独立测试，在自己创建的临时 Go module 中执行 `go mod tidy`、`go test -count=1 -race`、`go vet`。测试 overlay 不复制或替换正文实现；新的 Go 块没有登记时门禁失败。反例也编译，并由测试证明其描述的失败，而非将它误认为推荐代码。

临时 module 和测试子进程的 `TMPDIR` / `GOTMPDIR` 位于书仓 `.validation-tmp/`，因此本书产物保持在 `~/workspace` 的工作树内；该目录不入 Git。脚本只清理自己创建的 module，不删除目录中其他任务的文件。

不带章节参数时验证所有已登记章节；这不等于尚未登记的章节也已验证。每章发布前必须覆盖其全部 Go 块，依赖版本在 manifest 固定。真实服务、SQL 迁移和 Compose E2E 仍由配套代码仓验证，不能用此脚本代替。
