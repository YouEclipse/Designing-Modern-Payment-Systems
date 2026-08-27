# 第 22 章《从零设计 Global Payment Routing Platform》交付记录

## 章节契约

- 核心命题：综合架构必须把 SLO、容量单位、降级动作与 owner 绑定到可证伪故障旅程；100k TPS 是设计需求，不是当前实测结论。
- 演进线：单区域安全 SALE → v3 可解释选路与持久恢复 → 运营反馈 → 配置/多 Region/智能化目标架构。
- 本章代码基线：可运行支付核心已有 Router v3、Payment durable recovery 和 Gateway observation ledger；全平台系统 harness、Fallback/Health/Config/Global ownership 尚未实现。
- 本章不新增业务抽象，只统一容量模型、四条旅程和 Go/No-Go 证据范围。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：书稿与代码仓库当前均为 private；链接需获授权的 GitHub 身份。本交付不改变仓库可见性，也不把匿名不可访问解释为 CI 缺失。
- 当前 CI 负责真实 PostgreSQL integration、镜像、Compose readiness 与支付 E2E；100k/global/live 证据不在该范围。

## 主张—实现—验证映射

| status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | Router v3 在硬门禁后使用 Cost/Performance/Scoring，并给出稳定 `candidate_order` 与首次执行快照；这不是可执行的跨 Attempt `RoutePlanDecision` | `services/router/internal/decision/optimization.go`；`cost`/`performance`/`scoring` | optimization、golden contract、deterministic concurrent replay tests |
| `implemented` | Payment 在首个 Provider I/O 前固定 Context/Decision/Risk/Optimization 与 executable deadline | `services/payment/internal/postgres/routing_evidence.go`；迁移 `006` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten`；routing-authority recovery tests |
| `implemented` | UNKNOWN recovery 使用同 Operation/connection/key/payload，预算和 deadline 不重置 | `services/payment/internal/recovery/planner.go`；迁移 `005`；Gateway recovery | `TestRecoveryBudgetDeadlineAndIdentitySurviveWorkerRestart`；`TestResponseLossRecoveryQueriesThenResubmitsTheSameDurableOperation` |
| `implemented` | Provider 观察追加保存，冲突进入 review | Gateway `observations.go`；迁移 `0004`/`0005` | observation unit/PostgreSQL/append-only tests |
| `implemented` | 商户并发 confirm 仍只有一个持久资金路径 | Payment idempotency/store；Compose E2E | `TestTwentyConcurrentConfirmationsUseOneDurableFundsPath`；`TestPaymentVerticalSlice` |
| `teaching_model` | 容量必须拆分 logical SALE、完整 Router evaluation 与 guardrail-only revalidation；固定 Plan Fallback 不重做候选生成/评分 | 本章 `capacityplan` Go 包 | `scripts/go-examples/tests/chapter22` compile/race/vet；guardrail 单列、故障余量、恢复放大、溢出/NaN tests |
| `target_contract` | 目标 Fallback authority 是 Router-authored `RoutePlanDecision` + veto-only revalidation + Payment-authored `FallbackExecutionBinding`；当前 runtime 不存在 | 本章 authority contract；未来 Router/Payment API 与 successor transaction | chapter22 Go 模型只单列 guardrail work rate；不能作为 runtime 完成证据 |

## 图表、配图与书稿验证

- Mermaid 覆盖职责总图、四条验收旅程与固定 Route Plan 的资金路径。
- 本章 Go 容量模型由正文提取，以 chapter22 专属 tests 做 compile/race/vet。
- 本章没有新增 Ian 插图；综合系统图和故障时序图承担精确验收关系。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-6-case-study/global-payment-routing-platform.md`、`scripts/verify-mermaid.sh part-6-case-study/global-payment-routing-platform.md` 与 `git diff --check`；这些门禁只证明容量算术、authority contract 和图可重现。
- 当前 runtime 主张只能引用本 manifest 列出的窄代码证据，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；100k/global/live、Fallback executor 与 launch certification 均不在 CI 证明范围。

## 已实现边界

- 当前实现是 tokenized-card automatic-capture SALE 的 local/test 支付核心，不是 15 PSP、20 Payment Method、100k TPS 或真钱 live 平台。
- recovery policy 的 24 小时/32 次是仓库本地策略，不是任何生产 Provider 的幂等保留期承诺。
- Router v3 的 Fee/Performance Artifact 是 local/test evidence，不是生产报价或因果成功率。

## 未实现与后续

- 无正式 `RoutePlanDecision`、Router-owned guardrail-only revalidation、Payment-authored `FallbackExecutionBinding` 或 successor Attempt executor；当前 `candidate_order` 不能被夸大为这些 runtime。无动态 Provider Health 服务。
- 无第 22 章统一 `test/system`、`tools/capacity-plan`、`tools/launchcheck` 与目标环境容量报告。
- 无 Routing Configuration Platform、Experiment/Model runtime、独立 Reconciliation/Settlement、Multi-Region Intent ownership/fencing。
- 无真实 PSP sandbox certification、live credential/limit、PCI/安全运营或 Go/No-Go 签核。
