# 第 1 章《一笔支付是如何发生的》交付记录

## 章节契约

- 核心命题：一笔支付是一组可关联、不可被迟到消息抹掉的资金事实，不是一次 HTTP 调用的成功布尔值。
- 演进线：同步 SALE → 履约与认证异步 → 响应丢失和补偿 → Intent / Attempt / Operation / Evidence。
- 新增概念：Authorization、Capture、Clearing、Settlement 的阶段差异；业务投影与资金事实分离；UNKNOWN 和 integrity review。
- 后续范围：第 3 章精确执行身份、第 10～12 章恢复与幂等、第 17～18 章事实和经济闭环。本章不开放 manual capture、Void 或 Refund API。

## 配套实现

- 固定实现：[adbdf75589dcb7b9ae088aa9f92fde2c2f063aeb](https://github.com/YouEclipse/payments/commit/adbdf75589dcb7b9ae088aa9f92fde2c2f063aeb)。
- CI：[32919381068 — success](https://github.com/YouEclipse/payments/actions/runs/32919381068)。
- 切片：tokenized card automatic-capture SALE；真实 PostgreSQL 保存三层状态、幂等、inbox/outbox；Simulator 模拟外部 Provider，不处理真钱。

## 主张—实现—测试

下列工程路径均相对于以上固定实现；正文示例是便于阅读的最小模型，不冒充完整存储层。

| 主张 | 工程文件 | 验证 |
| --- | --- | --- |
| Authorization 不等于 captured success | `services/payment/internal/domain/model.go`；`services/provider-gateway/internal/gateway/service.go` | `TestSuccessfulAutomaticCaptureTransition`；`TestAuthorizedResponseDoesNotBecomeCapturedSuccess` |
| UNKNOWN 阻止第二次 Confirm，沿原 Operation 收敛 | Payment domain、app reconciler、Gateway service | `TestUnknownBlocksNewConfirmationAndConverges`；Compose `TestPaymentVerticalSlice` |
| 成功之后的冲突不能覆盖资金事实 | Payment domain、PostgreSQL store | `TestConflictingTerminalOutcomeRequiresReview`；`TestReviewRequiredIsNotIgnoredAfterTerminalSuccess` |
| 同事件 ID 的不同内容不是正常重复 | Payment inbox、Gateway webhook store | `TestDuplicateInboxEventRejectsChangedPayloadWithSameCorrelation`；正文 `TestProjectionRetainsSuccessAndFlagsConflicts` |
| 同步状态与幂等结果必须一起提交 | `services/payment/internal/postgres/store.go` | `TestApplyOutcomeCompletesConfirmationIdempotencyAtomically` |
| 生产不能开启模拟资金通道 | Gateway config / Provider Simulator | `TestProductionRejectsSimulatorProvider`；环境标准化测试 |

## 书稿验证

- 3 个 Go 块全部由 [可重复验证脚本](../scripts/go-examples/README.md) 从当前 Markdown 提取，编译、race、vet 通过；同步反例有测试证明其误分类 timeout 的问题。
- 4 个 Mermaid 用 `scripts/verify-mermaid.sh` 实际渲染通过。
- 1 张 Ian 风格图：`assets/payment-lifecycle-illustrations/01-success-is-not-one-fact.png`，1672×941，已经检查人物、标注和留白。
- 官方资料在正文就近引用；本地链接与资源存在性、`git diff --check` 已检查。
- 实现提交的 unit、race、vet、OpenAPI、真实 PostgreSQL integration、四镜像构建和 Compose E2E 全绿；代码远端 SHA 与上述提交一致。
- 书稿 revision 取包含本记录的 Git commit，不在文件中自引用尚未生成的书稿 SHA。

## 能力边界

这是可运行的支付核心与外部模拟器，不是 PSP sandbox 或 live 收款系统。示意状态图包含未来 manual-capture/refund 投影；现有 enum、不可达预留状态与数据库基数以附录 A 和固定代码为准。Ledger、结算、争议与真实 PSP 凭证流程不在此切片内。
