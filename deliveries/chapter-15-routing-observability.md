# 第 15 章《路由可观测性》交付记录

## 章节契约

- 核心命题：路由服务、支付服务商技术调用、支付资金结果与成熟经济结果必须使用不同统计单位、时间轴和所有者；观测系统不能获得资金写权限。
- 演进线：单一成功率 → 四类事实 → 有界维度与关联编号 → 不可重置年龄、告警和运行手册。
- 本章代码基线：已有持久派发与恢复时间锚点、网关追加式观察账本；完整遥测运行能力仍是设计型增量。
- 后续范围：实验因果判断、独立对账、结算实际值与全平台容量证据分别留给第 16～18、22 章。

## 能力状态

- `implemented`：固定实现中的不可重置派发与恢复时间锚点、追加式服务商观察与 `UNKNOWN` 原操作恢复。
- `teaching_model`：正文中的有限基数指标、追踪链接与四类观测事实 Go 模型。
- `target_contract`：共享遥测契约、OpenTelemetry 导出器、库存收集器、服务级指标仪表盘、告警投递与真实 PSP 可观测性。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- 持续集成任务：[33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：代码与书稿仓库当前均为私有仓库，链接需要获授权的 GitHub 身份；本交付不改变仓库可见性。
- 本章没有新增生产遥测服务；只把正文基线校准到已验证的资金证据能力。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 首次可能外发时间与恢复期限不能被查询或退避重置 | `services/payment/internal/postgres/migrations/005_payment_recovery.sql`；`services/provider-gateway/internal/gateway/migrations/0003_recovery_invariants.sql` | `TestRecoveryBudgetDeadlineAndIdentitySurviveWorkerRestart`；`TestPostgresGatewayOperationRecoveryInvariants` |
| `implemented` | 服务商观察追加保存，状态归约不能抹掉原证据 | `services/provider-gateway/internal/gateway/observations.go`；迁移 `0004` 与 `0005` | `TestPostgresProviderObservationLedgerFencesStaleGenerationAndIsImmutable`；`TestProviderObservationLedgerCoversSaleQueryAndWebhook` |
| `teaching_model` | 有限维度、指标单位与追踪链接必须通过结构校验 | 本章 `telemetry` Go 包 | `scripts/go-examples/tests/chapter15` 编译、竞态检测与静态检查；有限维度与追踪链接测试 |
| `implemented` | `UNKNOWN` 恢复始终沿原连接与资金操作，并使用持久服务商幂等键，不重新选路 | `services/payment/internal/recovery/planner.go`；`services/payment/internal/app/recovery_execution.go` | `TestResponseLossRecoveryQueriesThenResubmitsTheSameDurableOperation`；Compose 响应丢失端到端测试 |
| `target_contract` | 未来遥测与告警不得授权换路或发起资金操作 | 未来遥测与支付权威边界合同 | 当前无告警投递与资金执行集成；正文权威边界 |

## 图表与书稿验证

- 5 个 Mermaid 用于四层事实、同步与异步追踪、双时间轴和 `UNKNOWN` 运行手册；不以装饰图替代精确关系。
- 本章 4 个 Go 示例包由正文提取，使用 `scripts/go-examples/tests/chapter15` 做编译、竞态检测和静态检查。
- 本章没有新增 Ian 插图；原因是时间轴、因果链接与告警权限用 Mermaid 更精确。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-4-operating-at-scale/observability.md`、`scripts/verify-mermaid.sh part-4-operating-at-scale/observability.md` 与 `git diff --check`；这些门禁只证明正文示例和图可重现。
- 当前运行能力主张只能引用交付记录中列出的代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色持续集成任务固定后发布；不得把本书遥测教学包写成已部署服务。

## 未实现与能力边界

- 尚无共享 `internal/telemetry`、OpenTelemetry 导出器、`UNKNOWN` 与发件箱库存收集器、完整服务级指标仪表盘与告警投递。
- 固定提交中的三个 HTTP 处理器各定义了仅含进程存活仪表值的 `/metrics`，但尚无直接端点回归测试；它不计作已验证的业务可观测能力，更不代表业务服务级指标。
- `created_at` 与 `first_dispatched_at` 可提供不可重置的保守年龄锚点；区分多轮未决区间的显式 `first_unresolved_at` 尚未实现。
- 观察账本不是对账库存、财务事实或指标后端；没有真实 PSP 生产可观测性声明。
