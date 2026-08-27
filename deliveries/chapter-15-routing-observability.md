# 第 15 章《Routing Observability》交付记录

## 章节契约

- 核心命题：Router 服务、Provider 技术调用、Payment 资金结果与成熟经济结果必须使用不同统计单位、时间轴和 owner；观测系统不能获得资金写权限。
- 演进线：单一成功率 → 四类事实 → 有界维度与关联 ID → 不可重置年龄、告警和 Runbook。
- 本章代码基线：已有 durable dispatch/recovery 时间锚点与 Gateway 追加式 observation ledger；完整 telemetry runtime 仍是设计型增量。
- 后续范围：实验因果判断、独立 Reconciliation、Settlement Actual 与全平台容量证据分别留给第 16～18、22 章。

## Status

- `implemented`：固定实现中的不可重置 dispatch/recovery 时间锚点、追加式 Provider observation 与 UNKNOWN 原操作恢复。
- `teaching_model`：正文中的有限基数 metric、trace link 与四类观测事实 Go 模型。
- `target_contract`：共享 telemetry 契约、OpenTelemetry exporter、inventory collector、SLI Dashboard、告警投递与真实 PSP 可观测性。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：代码与书稿仓库当前均为 private，链接需要获授权的 GitHub 身份；本交付不改变仓库可见性。
- 本章没有新增生产 telemetry 服务；只把正文基线校准到已验证的资金证据能力。

## 主张—实现—验证映射

| Status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 首次可能外发时间与恢复期限不能被 Query/backoff 重置 | `services/payment/internal/postgres/migrations/005_payment_recovery.sql`；`services/provider-gateway/internal/gateway/migrations/0003_recovery_invariants.sql` | `TestRecoveryBudgetDeadlineAndIdentitySurviveWorkerRestart`；`TestPostgresGatewayOperationRecoveryInvariants` |
| `implemented` | Provider 观察追加保存，状态归约不能抹掉原证据 | `services/provider-gateway/internal/gateway/observations.go`；迁移 `0004`/`0005` | `TestPostgresProviderObservationLedgerFencesStaleGenerationAndIsImmutable`；`TestProviderObservationLedgerCoversSaleQueryAndWebhook` |
| `teaching_model` | 有限维度、metric unit 与 trace link 必须通过结构校验 | 本章 `telemetry` Go 包 | `scripts/go-examples/tests/chapter15` compile/race/vet；有限维度与 trace-link 测试 |
| `implemented` | UNKNOWN 恢复始终沿原 connection/operation 与持久 Provider key，不重新选路 | `services/payment/internal/recovery/planner.go`；`services/payment/internal/app/recovery_execution.go` | `TestResponseLossRecoveryQueriesThenResubmitsTheSameDurableOperation`；Compose response-loss E2E |
| `target_contract` | 未来 telemetry/告警不得授权换路或发起资金操作 | 未来 telemetry 与 Payment authority 合同 | 当前无告警投递与资金执行集成；正文 authority 边界 |

## 图表与书稿验证

- 5 个 Mermaid 用于四层事实、同步/异步 trace、双时间轴和 UNKNOWN Runbook；不以装饰图替代精确关系。
- 本章 4 个 Go package 由正文提取，使用 `scripts/go-examples/tests/chapter15` 做 compile/race/vet。
- 本章没有新增 Ian 插图；原因是时间轴、因果 link 与告警权限用 Mermaid 更精确。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-4-operating-at-scale/observability.md`、`scripts/verify-mermaid.sh part-4-operating-at-scale/observability.md` 与 `git diff --check`；这些门禁只证明正文示例和图可重现。
- 当前 runtime 主张只能引用 manifest 中列出的代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；不得把本书 telemetry 教学包写成已部署服务。

## 未实现与能力边界

- 尚无共享 `internal/telemetry`、OpenTelemetry exporter、UNKNOWN/outbox inventory collector、完整 SLI Dashboard 与告警投递。
- 固定 SHA 的三个 HTTP handler 各定义了仅含 process-up gauge 的 `/metrics`，但尚无直接 endpoint regression；它不计作已验证的业务可观测能力，更不代表业务 SLI。
- `created_at`/`first_dispatched_at` 可提供不可重置的保守年龄锚点；区分多轮未决区间的显式 `first_unresolved_at` 尚未实现。
- Observation ledger 不是 Reconciliation inventory、财务事实或 Metric backend；没有真实 PSP/live 可观测性声明。
