# 第 10 章《超时、重试与未知状态》交付记录

## 章节契约

- 核心命题：超时只证明观测缺失；恢复必须固定原资金操作、精确连接修订、服务商幂等键与请求载荷，并受持久化截止时间、预算与证据门槛约束。
- 演进线：通用 POST 重试 → `UNKNOWN` 或可能已执行 → 查询原资金操作 → 有类型的确定性未执行证据 → 带代数的恢复认领与原键重提。
- 本章代码增量：用于 `SALE`、请款模式为 `automatic` 的纯恢复规划器，PostgreSQL 截止时间、预算与认领，首次派发标记、结构化未执行证明与追加式服务商观察。
- 后续范围：真实 PSP 查询、幂等与未找到合同，生产调度器与容量运维，人工处置控制面仍未实现。

## 配套实现

- 代码仓库：`YouEclipse/payments`
- 实现提交：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- 持续集成任务：[33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 纵向切片：响应丢失后只查询或用原键恢复原资金操作；普通 404 不产生第二次 POST，只有绑定完整身份且未过期的结构化未执行证明才允许原键重提。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `teaching_model` | 正文规划器用精简的 `OriginalCommand` 与 `Snapshot` 展示查询、等待、原键重提、停止四个动作；它不表示完整运行时模式 | 本章 `package recovery` Go 代码块 | `scripts/go-examples/tests/chapter10/recovery_test.go`；`TestRecoveryActionMatrixDoesNotRerouteOrMutateIdentity`、`TestRecoveryRejectsMissingIdentityAndUnrecognizedEvidence` |
| `implemented` | 真实规划器固定资金操作、支付尝试、连接修订、命名空间、服务商、幂等键、载荷哈希，并结合认领依据、首次派发、截止时间、预算、退避、租约与有类型的未执行证明；普通 `UNKNOWN` 不进入后备路由 | `services/payment/internal/recovery/planner.go`；`services/payment/internal/app/recovery_execution.go`；`services/payment/internal/postgres/recovery.go` | `TestRecoveryActionTable`、`TestUnprovenOrMalformedAbsenceCannotBecomeResubmit`、`TestAbsenceEvidenceRequiresWholeIdentityAndHalfOpenValidity` |
| `implemented` | 截止时间、预算、首次派发、认领代数与身份跨重启不变 | `services/payment/internal/postgres/migrations/005_payment_recovery.sql`；`services/payment/internal/postgres/recovery.go` | `TestRecoveryBudgetDeadlineAndIdentitySurviveWorkerRestart`、`TestCallBudgetAndJitterCannotExtendDeadline` |
| `implemented` | 首次网关输入输出与已派发操作的恢复使用不同认领依据，执行权限不会被租约误解 | `services/payment/internal/app/recovery_execution.go`；`services/payment/internal/postgres/recovery.go` | `TestRoutingAuthorityExpiryAfterClaimStopsBeforeFirstGatewayIO`、`TestExpiredRoutingAuthorityDoesNotBlockAlreadyDispatchedSameKeyRecovery` |
| `implemented` | 确定性未执行证据必须绑定资金操作、支付尝试、连接修订、命名空间、服务商、幂等键、载荷与半开有效期 | `services/payment/internal/clients/gateway_recovery.go`；`services/provider-gateway/internal/gateway/recovery.go` | `TestAbsenceEvidenceRequiresWholeIdentityAndHalfOpenValidity`、`TestStructuredRecoveryInspectionBindsTheWholeV2Command` |
| `implemented` | 普通 404 继续查询原资金操作，不会变成第二次服务商 POST | `services/provider-gateway/internal/gateway/service.go`；`services/payment/internal/app/recovery_execution.go` | `TestUnstructuredProvider404RemainsUnknownWithoutResubmit`、`TestRecoveryOrdinary404NeverBecomesASecondPost` |
| `implemented` | 销售、查询与回调通知证据只追加，迟到或冲突证据不覆盖已请款事实 | `services/provider-gateway/internal/gateway/observations.go`；迁移 `0004` 与 `0005` | `TestPostgresProviderObservationLedgerFencesStaleGenerationAndIsImmutable`、`TestPostgresWebhookObservationIsAtomicWithReducerAndRetainsConflictingDelivery` |
| `implemented` | 支付服务用进程时钟判断截止时间与租约，并在就绪检查、新确认、恢复认领以及紧邻首次或重提资金输入输出前，以 PostgreSQL `clock_timestamp()` 见证；超出 `PAYMENT_MAX_CLOCK_SKEW` 时拒绝执行 | `services/payment/internal/config/config.go`；`services/payment/internal/postgres/store.go`；`services/payment/internal/app/service.go`、`recovery_execution.go` | `TestDatabaseClockMustStayInsideTheMeasuredProcessWindow`、`TestProductionRequiresAnExplicitPositiveClockSkewBudget`；`docs/runbooks/clock-safety.md` |
| `implemented` | 响应丢失最终收敛且全程只有一条资金路径 | `test/e2e/payment_flow_test.go` | `TestPaymentVerticalSlice` 的 `timeout_then_succeeded` 场景 |

## 图表与配图

- 4 个 Mermaid：响应丢失恢复序列、三层截止时间、规划器证据流、恢复状态机。
- Ian 正文配图：`assets/timeout-and-retry-illustrations/01-unknown-follow-the-original-operation.png`，1672×941；表达 `UNKNOWN` 只能沿原资金操作取证。
- 3 个 Go 代码块由正文提取，配套第 10 章测试覆盖不安全重试、观察归一化与恢复规划器边界。

## 发布验证

- 书稿侧执行章节 Go 编译、竞态检测、静态检查、Mermaid 实际渲染、本地链接、图片路径与尺寸检查，以及 `git diff --check`。
- 代码侧单元测试、竞态检测、静态检查、OpenAPI、真实 PostgreSQL 集成测试与 Compose 端到端测试的发布证据，以本记录固定的持续集成任务为准；该任务必须成功。
- 模拟器只验证故障与证据合同；系统没有真实 PSP 沙箱或生产资金能力。

## 未实现与后续章节

- 未接入真实 PSP 的权威查询、索引一致性窗口、未执行证明能力、幂等键保留期与官方错误分类。
- 现有工作进程与认领机制是可运行核心，不等于生产级调度器、容量隔离、告警和运营手册。
- PostgreSQL 目前只是支付服务进程时钟的部署见证，不是所有截止时间与租约的数据库时间事务权威；生产必须监控 NTP 偏移、支付服务→PostgreSQL 往返时间与见证失败，路由服务与网关时钟仍需独立保护。长期可把恢复认领与租约改为数据库时间权威，不能通过放宽 `PAYMENT_MAX_CLOCK_SKEW` 掩盖不安全时钟。
