# 第 22 章《从零设计全球支付路由平台》交付记录

## 章节契约

- 核心命题：综合架构必须把 SLO、容量单位、降级动作与负责人绑定到可证伪故障旅程；10 万 TPS 是设计需求，不是当前实测结论。
- 演进线：单区域安全 `SALE` → v3 可解释选路与持久恢复 → 运营反馈 → 配置、多区域与智能化目标架构。
- 本章代码基线：可运行支付核心已有路由服务 v3、支付持久恢复和网关观察账本；全平台系统编排器、后备切换、健康度、配置与全球归属尚未实现。
- 本章不新增业务抽象，只统一容量模型、四条旅程和上线/不上线证据范围。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：书稿与代码仓库当前均为私有；链接需获授权的 GitHub 身份。本交付不改变仓库可见性，也不把匿名不可访问解释为 CI 缺失。
- 当前 CI 负责真实 PostgreSQL 集成、镜像、Compose 就绪检查与支付端到端测试；10 万、全球化与生产证据不在该范围。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 路由服务 v3 在硬门禁后使用成本、性能与评分，并给出稳定的 `candidate_order` 与首次执行快照；这不是可执行的跨支付尝试 `RoutePlanDecision` | `services/router/internal/decision/optimization.go`；`cost`、`performance`、`scoring` | 优化、黄金契约与确定性并发重放测试 |
| `implemented` | 支付服务在首次服务商 I/O 前固定上下文、决策、风险、优化与可执行截止时间 | `services/payment/internal/postgres/routing_evidence.go`；迁移 `006` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten`；路由权限恢复测试 |
| `implemented` | `UNKNOWN` 恢复使用同一操作、连接、键与载荷，预算和截止时间不重置 | `services/payment/internal/recovery/planner.go`；迁移 `005`；网关恢复 | `TestRecoveryBudgetDeadlineAndIdentitySurviveWorkerRestart`；`TestResponseLossRecoveryQueriesThenResubmitsTheSameDurableOperation` |
| `implemented` | 服务商观察追加保存，冲突进入人工复核 | 网关 `observations.go`；迁移 `0004`/`0005` | 观察单元测试、PostgreSQL 测试与仅追加测试 |
| `implemented` | 商户并发确认仍只有一个持久资金路径 | 支付幂等与存储；Compose 端到端测试 | `TestTwentyConcurrentConfirmationsUseOneDurableFundsPath`；`TestPaymentVerticalSlice` |
| `teaching_model` | 容量必须拆分逻辑 `SALE`、完整路由求值与仅护栏复核；固定计划中的后备切换不重做候选生成或评分 | 本章 `capacityplan` Go 包 | `scripts/go-examples/tests/chapter22` 编译、竞态与静态检查；护栏单列、故障余量、恢复放大、溢出与 NaN 测试 |
| `target_contract` | 目标后备切换权限由路由服务生成的 `RoutePlanDecision`、仅否决复核与支付服务生成的 `FallbackExecutionBinding` 共同构成；当前运行时不存在 | 本章权限合同；未来路由与支付 API 及后继事务 | 第 22 章 Go 模型只单列护栏工作速率；不能作为运行时完成证据 |

## 图表、配图与书稿验证

- Mermaid 覆盖职责总图、四条验收旅程与固定路由计划的资金路径。
- 本章 Go 容量模型由正文提取，以第 22 章专属测试做编译、竞态与静态检查。
- 本章没有新增 Ian 插图；综合系统图和故障时序图承担精确验收关系。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-6-case-study/global-payment-routing-platform.md`、`scripts/verify-mermaid.sh part-6-case-study/global-payment-routing-platform.md` 与 `git diff --check`；这些门禁只证明容量算术、权限合同和图可重现。
- 当前运行时主张只能引用本交付清单列出的窄代码证据，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；10 万、全球与生产能力、后备执行器及上线认证均不在 CI 证明范围。

## 已实现边界

- 当前实现是令牌化银行卡自动请款 `SALE` 的本地与测试支付核心，不是覆盖 15 家 PSP、20 种支付方式、10 万 TPS 或真钱的生产平台。
- 恢复策略的 24 小时与 32 次是仓库本地策略，不是任何生产服务商的幂等保留期承诺。
- 路由服务 v3 的费用与性能制品是本地和测试证据，不是生产报价或因果成功率。

## 未实现与后续

- 无正式 `RoutePlanDecision`、路由服务拥有的仅护栏复核、支付服务生成的 `FallbackExecutionBinding` 或后继支付尝试执行器；当前 `candidate_order` 不能被夸大为这些运行时。无动态服务商健康度服务。
- 无第 22 章统一 `test/system`、`tools/capacity-plan`、`tools/launchcheck` 与目标环境容量报告。
- 无路由配置平台、实验与模型运行时、独立对账与结算、多区域支付意图归属与隔离机制。
- 无真实 PSP 沙箱认证、生产凭证与限额、PCI 与安全运营或上线签核。
