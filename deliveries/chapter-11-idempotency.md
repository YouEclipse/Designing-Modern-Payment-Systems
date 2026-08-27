# 第 11 章《Idempotency 与支付尝试》交付记录

## 章节契约

- 核心命题：幂等是持久化业务身份、参数不变性与并发所有权；Merchant key、Operation ID、Provider key、Provider reference 不可混用。
- 演进线：Redis response cache → PostgreSQL command record → typed/versioned fingerprint → Serializable winner → Provider/Gateway identity 与 inbox 边界。
- 本章代码增量：20 路并发 Confirm 的单 owner、进程重建 Replay、`command_sha256.v1` fingerprint version、Gateway v2 typed request hash。
- 后续范围：Payment inbox canonical payload hash/version、真实 Provider retention metadata、cleanup/archive lifecycle 与跨 Region owner。

## 配套实现

- Repository：`YouEclipse/payments`
- Implementation commit：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- CI run：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 纵向切片：20 个相同 Confirm 共享一个 Intent、Attempt、Payment Operation 与一次 Gateway execute；重建 Service 后仍从 PostgreSQL Replay。

## 主张—实现—验证映射

| status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | Merchant key 绑定 operation 与 typed payload；未知 fingerprint version fail closed | `services/payment/internal/idempotency/fingerprint.go`；migration `005_payment_recovery.sql` | `TestFingerprintBindsOperationAndPayload`、`TestScanIdempotencyRejectsReusedKey` |
| `implemented` | Serializable transaction/唯一约束选出一个 Confirm owner | `services/payment/internal/postgres/store.go`；`transaction_retry.go` | `TestTwentyConcurrentConfirmationsUseOneDurableFundsPath` |
| `implemented` | 20 个 caller 只有一个 Attempt/Operation/Gateway execute，重启后同资源 Replay | `services/payment/internal/app/idempotency_integration_test.go` | `TestTwentyConcurrentConfirmationsUseOneDurableFundsPath` |
| `implemented` | Provider request hash 绑定完整 typed execution identity | `services/provider-gateway/internal/gateway/service.go`；Payment `clients/gateway.go` | `TestOperationHashV2UsesUnambiguousSemanticEncoding`、`TestOperationHashV2GoldenPaymentContract`、`TestRecoveryRejectsDifferentPayloadForSameOperation` |
| `implemented` | Provider key/reference/event 的冲突唯一性按 `provider_namespace_id`，不是按 connection；验签与执行仍固定精确 connection/revision | Gateway migrations/store/connection registry | `TestGatewayMigrationsCreateFreshIdentitySchema`、`TestPostgresStoreNamespaceScopedWebhookAndReferenceCollisions`、connection drift tests |
| `implemented` | captured 与 idempotency completion 原子提交，响应丢失不重开执行权 | `services/payment/internal/postgres/store.go` | `TestApplyOutcomeCompletesConfirmationIdempotencyAtomically` |
| `implemented` | inbox 重复只应用一次，correlation/payload 冲突保留并进入复核 | Payment/Gateway store 与 observation ledger | `TestDuplicateInboxEventRejectsChangedPayloadWithSameCorrelation`、`TestWebhookEventIDCannotBeReusedForAnotherOperation` |

## 图表与配图

- 5 个 Mermaid：四层身份、并发 owner、crash/replay、领域 ER 与 inbox 收敛。
- 本章未添加 Ian 插图；身份/事务关系由 Mermaid 与对照表表达更精确。
- 4 个 Go 块由正文提取，配套 chapter11 tests 覆盖 cache 反例、typed fingerprint、Replay plan 与 Provider command digest。

## 发布验证

- 书稿侧执行章节 Go compile/race/vet、Mermaid 实际渲染、本地链接与 `git diff --check`。
- 代码侧真实 PostgreSQL 并发测试、unit/race/vet、OpenAPI 与 Compose E2E 的最终证据以固定 CI run 为准；该 run 必须为 success。
- 当前能力仅覆盖 tokenized card `SALE/automatic`，不代表 Capture/Void/Refund 或 live PSP 已实现。

## 未实现与后续章节

- Payment inbox 尚无显式 canonical payload hash 与算法版本；Gateway observation 的 hash 不能替代 Payment-owned inbox 合同。
- 真实 Provider key 作用域/retention metadata、记录清理/归档/Legal Hold 与过期后不重开资金路径的测试仍缺失。
