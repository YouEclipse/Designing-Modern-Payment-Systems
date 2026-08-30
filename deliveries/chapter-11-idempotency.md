# 第 11 章《幂等性与支付尝试》交付记录

## 章节契约

- 核心命题：幂等是持久化业务身份、参数不变性与并发所有权；商户幂等键、资金操作编号、服务商幂等键和服务商引用不可混用。
- 演进线：Redis 响应缓存 → PostgreSQL 命令记录 → 有类型、带版本的指纹 → 可串行化事务选出胜出项 → 服务商与网关身份、收件箱边界。
- 本章代码增量：20 路并发确认的单一所有者、进程重建后的重放、`command_sha256.v1` 指纹版本、网关 v2 有类型的请求哈希。
- 后续范围：支付收件箱的规范载荷哈希与版本、真实服务商保留期元数据、清理与归档生命周期、跨区域所有者。

## 配套实现

- 代码仓库：`YouEclipse/payments`
- 实现提交：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- 持续集成任务：[33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 纵向切片：20 个相同确认共享一个支付意图、支付尝试、支付资金操作与一次网关执行；重建服务后仍从 PostgreSQL 重放。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 商户幂等键绑定资金操作与有类型的载荷；指纹版本未知时拒绝执行 | `services/payment/internal/idempotency/fingerprint.go`；迁移 `005_payment_recovery.sql` | `TestFingerprintBindsOperationAndPayload`、`TestScanIdempotencyRejectsReusedKey` |
| `implemented` | 可串行化事务与唯一约束选出一个确认所有者 | `services/payment/internal/postgres/store.go`；`transaction_retry.go` | `TestTwentyConcurrentConfirmationsUseOneDurableFundsPath` |
| `implemented` | 20 个调用方只有一个支付尝试、一项资金操作和一次网关执行；重启后重放同一资源 | `services/payment/internal/app/idempotency_integration_test.go` | `TestTwentyConcurrentConfirmationsUseOneDurableFundsPath` |
| `implemented` | 服务商请求哈希绑定完整且有类型的执行身份 | `services/provider-gateway/internal/gateway/service.go`；支付侧 `clients/gateway.go` | `TestOperationHashV2UsesUnambiguousSemanticEncoding`、`TestOperationHashV2GoldenPaymentContract`、`TestRecoveryRejectsDifferentPayloadForSameOperation` |
| `implemented` | 服务商幂等键、引用与事件的冲突唯一性按 `provider_namespace_id`，不是按连接；验签与执行仍固定精确连接和修订 | 网关迁移、存储与连接注册表 | `TestGatewayMigrationsCreateFreshIdentitySchema`、`TestPostgresStoreNamespaceScopedWebhookAndReferenceCollisions`、连接漂移测试 |
| `implemented` | 已请款状态与幂等完成记录原子提交，响应丢失不重开执行权 | `services/payment/internal/postgres/store.go` | `TestApplyOutcomeCompletesConfirmationIdempotencyAtomically` |
| `implemented` | 收件箱重复事件只应用一次，关联编号与载荷冲突会保留并进入复核 | 支付服务与网关存储、观察账本 | `TestDuplicateInboxEventRejectsChangedPayloadWithSameCorrelation`、`TestWebhookEventIDCannotBeReusedForAnotherOperation` |

## 图表与配图

- 5 个 Mermaid：四层身份、并发所有者、崩溃与重放、领域实体关系与收件箱收敛。
- 本章未添加 Ian 插图；身份与事务关系由 Mermaid 与对照表表达更精确。
- 4 个 Go 代码块由正文提取，配套第 11 章测试覆盖缓存反例、有类型的指纹、重放计划与服务商命令摘要。

## 发布验证

- 书稿侧执行章节 Go 编译、竞态检测、静态检查、Mermaid 实际渲染、本地链接与 `git diff --check`。
- 代码侧真实 PostgreSQL 并发测试、单元测试、竞态检测、静态检查、OpenAPI 与 Compose 端到端测试的最终证据，以固定持续集成任务为准；该任务必须成功。
- 当前能力仅覆盖令牌化银行卡的 `SALE` 操作与 `automatic` 请款模式，不代表请款、撤销、退款或真实生产 PSP 已实现。

## 未实现与后续章节

- 支付收件箱尚无显式规范载荷哈希与算法版本；网关观察记录的哈希不能替代支付服务所有的收件箱合同。
- 真实服务商幂等键的作用域与保留期元数据、记录清理、归档、诉讼保留，以及过期后不重开资金路径的测试仍缺失。
