# 第 10 章《Timeout、Retry 与 Unknown State》交付记录

## 章节契约

- 核心命题：超时只证明观测缺失；恢复必须固定原 Operation、精确 connection revision、Provider key/payload，并受持久化 deadline/budget 与证据门槛约束。
- 演进线：通用 POST retry → UNKNOWN/possible → 原 Operation Query → typed conclusive absence → 有代数的 Recovery claim 与 Same-Key Resubmit。
- 本章代码增量：`SALE/automatic` 的纯 RecoveryPlanner、PostgreSQL deadline/budget/claim、首次 dispatch marker、结构化 absence proof 与 append-only Provider Observation。
- 后续范围：真实 PSP query/idempotency/not-found 合同、生产 scheduler/容量运维、人工处置控制面仍未实现。

## 配套实现

- Repository：`YouEclipse/payments`
- Implementation commit：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- CI run：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 纵向切片：响应丢失后只查询/同键恢复原 Operation；普通 404 不产生第二次 POST，只有绑定完整 identity 且未过期的 structured absence proof 才允许 Same-Key Resubmit。

## 主张—实现—验证映射

| status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `teaching_model` | 正文 Planner 用精简 `OriginalCommand`/`Snapshot` 展示 Query/Wait/Same-Key Resubmit/Stop 四动作；它不表示完整 runtime schema | 本章 `package recovery` Go block | `scripts/go-examples/tests/chapter10/recovery_test.go`；`TestRecoveryActionMatrixDoesNotRerouteOrMutateIdentity`、`TestRecoveryRejectsMissingIdentityAndUnrecognizedEvidence` |
| `implemented` | 真实 Planner 固定 operation/attempt/connection revision/namespace/provider/key/payload hash，并结合 ClaimBasis、first dispatch、deadline、budget、backoff、lease 与 typed absence；普通 UNKNOWN 不进入 Fallback | `services/payment/internal/recovery/planner.go`；`services/payment/internal/app/recovery_execution.go`；`services/payment/internal/postgres/recovery.go` | `TestRecoveryActionTable`、`TestUnprovenOrMalformedAbsenceCannotBecomeResubmit`、`TestAbsenceEvidenceRequiresWholeIdentityAndHalfOpenValidity` |
| `implemented` | deadline、budget、first dispatch、claim generation 与 identity 跨重启不变 | `services/payment/internal/postgres/migrations/005_payment_recovery.sql`；`services/payment/internal/postgres/recovery.go` | `TestRecoveryBudgetDeadlineAndIdentitySurviveWorkerRestart`、`TestCallBudgetAndJitterCannotExtendDeadline` |
| `implemented` | 首次 Gateway I/O 与已 dispatch recovery 使用不同 claim basis，执行权限不会被 lease 误解 | `services/payment/internal/app/recovery_execution.go`；`services/payment/internal/postgres/recovery.go` | `TestRoutingAuthorityExpiryAfterClaimStopsBeforeFirstGatewayIO`、`TestExpiredRoutingAuthorityDoesNotBlockAlreadyDispatchedSameKeyRecovery` |
| `implemented` | conclusive absence 必须绑定 operation/attempt/connection revision/namespace/provider/key/payload 与半开有效期 | `services/payment/internal/clients/gateway_recovery.go`；`services/provider-gateway/internal/gateway/recovery.go` | `TestAbsenceEvidenceRequiresWholeIdentityAndHalfOpenValidity`、`TestStructuredRecoveryInspectionBindsTheWholeV2Command` |
| `implemented` | 普通 404 继续 Query 原 Operation，不会变成第二次 Provider POST | `services/provider-gateway/internal/gateway/service.go`；`services/payment/internal/app/recovery_execution.go` | `TestUnstructuredProvider404RemainsUnknownWithoutResubmit`、`TestRecoveryOrdinary404NeverBecomesASecondPost` |
| `implemented` | sale/query/webhook 证据 append-only，迟到或冲突证据不覆盖 captured 事实 | `services/provider-gateway/internal/gateway/observations.go`；migrations `0004`/`0005` | `TestPostgresProviderObservationLedgerFencesStaleGenerationAndIsImmutable`、`TestPostgresWebhookObservationIsAtomicWithReducerAndRetainsConflictingDelivery` |
| `implemented` | Payment 用进程时钟判断 deadline/lease，并在 readiness、新 Confirm、recovery claim 与紧邻首次/重提资金 I/O 前以 PostgreSQL `clock_timestamp()` 见证；超出 `PAYMENT_MAX_CLOCK_SKEW` 时 fail closed | `services/payment/internal/config/config.go`；`services/payment/internal/postgres/store.go`；`services/payment/internal/app/service.go`、`recovery_execution.go` | `TestDatabaseClockMustStayInsideTheMeasuredProcessWindow`、`TestProductionRequiresAnExplicitPositiveClockSkewBudget`；`docs/runbooks/clock-safety.md` |
| `implemented` | 响应丢失最终收敛且全程只有一条资金路径 | `test/e2e/payment_flow_test.go` | `TestPaymentVerticalSlice` 的 `timeout_then_succeeded` 场景 |

## 图表与配图

- 4 个 Mermaid：响应丢失恢复序列、三层 deadline、Planner 证据流、Recovery 状态机。
- Ian 正文配图：`assets/timeout-and-retry-illustrations/01-unknown-follow-the-original-operation.png`，1672×941；表达 UNKNOWN 只能沿原 Operation 取证。
- 3 个 Go 块由正文提取，配套 chapter10 tests 覆盖 unsafe retry、Observation normalization 与 RecoveryPlanner 边界。

## 发布验证

- 书稿侧执行章节 Go compile/race/vet、Mermaid 实际渲染、本地链接、图片路径/尺寸与 `git diff --check`。
- 代码侧 unit、race、vet、OpenAPI、真实 PostgreSQL integration 与 Compose E2E 的发布证据以本记录固定的 CI run 为准；该 run 必须为 success。
- Simulator 只验证故障与证据合同；没有真实 PSP sandbox/live 资金能力。

## 未实现与后续章节

- 未接入真实 PSP 的权威 Query、索引一致性窗口、absence capability、key retention 与官方错误分类。
- 现有 worker/claim 是可运行核心，不等于生产级 scheduler、容量隔离、告警和运营 runbook。
- PostgreSQL 目前只是 Payment 进程时钟的部署见证，不是所有 deadline/lease 的数据库时间事务权威；生产必须监控 NTP offset、Payment→PostgreSQL RTT 与见证失败，Router/Gateway 时钟仍需独立保护。长期可把 recovery claim/lease 改为数据库时间权威，不能通过放宽 `PAYMENT_MAX_CLOCK_SKEW` 掩盖不安全时钟。
