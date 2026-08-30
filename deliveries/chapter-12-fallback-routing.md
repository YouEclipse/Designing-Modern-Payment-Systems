# 第 12 章《后备路由》交付记录

## 章节契约

- 核心命题：只有旧支付尝试满足“已终止 + 无副作用 + 原因在允许清单中”，且所有护栏都通过时，支付服务才可沿固定路由计划创建一个新支付尝试。
- 演进线：遇错换 PSP → 确定性与原因双闸门 → 固定候选顺序 → 新身份后继项 → 并发唯一事务。
- 当前代码基线：v3 完整上下文、决策、风险、优化、顺序与截止时间的原子快照，以及 `UNKNOWN` 默认阻断。
- 明确设计项：路由服务编写的 `RoutePlanDecision`、路由服务所有且只能否决的护栏复检、支付服务编写的 `FallbackExecutionBinding`、后继事务与跨支付服务商端到端测试尚未实现。

## 配套实现

- 代码仓库：`YouEclipse/payments`
- 实现提交：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- 持续集成任务：[33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 本章纵向切片：本轮只交付后备路由的安全前置条件与可编译教学模型，不宣称代码仓库已经能创建支付尝试 2。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | `UNKNOWN` 或可能已执行时只恢复原资金操作，默认禁止新建支付尝试 | `services/payment/internal/domain/model.go`；`services/payment/internal/recovery/planner.go` | `TestUnknownBlocksNewConfirmationAndConverges`、`TestRecoveryActionTable` |
| `implemented` | v3 上下文、风险、优化、候选顺序与截止时间在首次支付尝试事务中原子固定 | `services/payment/internal/postgres/routing_evidence.go`；迁移 `006_risk_routing_evidence.sql`；`store.go` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten`、`TestRoutingEvidenceBindsLockedIntentAndSelectedExecutionIdentity` |
| `implemented` | 风险阻断、人工复核与要求 3DS 都不能产生可执行候选通道 | 路由服务 v3 与严格支付客户端 | `TestRouterV3RiskActionsCannotCreateAnExecutableCandidate`、`TestV3HTTPSuccessMatchesSharedContractGolden` |
| `teaching_model` | 后备路由教学规划器只消费路由服务编写的 `RoutePlanDecision` 固定下一项；复检只能否决，不能重排或续期 | 本章 `package fallback` Go 代码块 | `scripts/go-examples/tests/chapter12/fallback_test.go` |
| `target_contract` | 获准后继项由支付服务写入 `FallbackExecutionBinding`，引用原计划、前序支付尝试、固定下一项与新鲜护栏证据，不冒充路由决策 | 本章权威边界合同；未来支付后继事务 | 正文中的权威、过期与否决故障模式；当前没有运行时测试 |
| `teaching_model` | 后继项必须使用新的支付尝试、资金操作和幂等键，不能改写旧身份 | 本章 `package successor` Go 代码块 | `scripts/go-examples/tests/chapter12/successor_test.go` |
| `implemented` | 两个确认结果不覆盖已请款事实，业务成功与完整性复核分轴 | `services/payment/internal/domain/model.go` | `TestConflictingTerminalOutcomeRequiresReview`、`TestManualReviewStopsAutomaticRecoveryAndPreservesFunds` |

## 图表与配图

- 4 个 Mermaid：误切换双扣、支付尝试链、规划器闸门、并发后继事务。
- 本章未添加 Ian 插图；资金路径与门禁更适合精确流程图。
- 3 个 Go 代码块由正文提取，第 12 章测试覆盖朴素反例、规划器与后继身份。

## 发布验证

- 书稿侧执行章节 Go 编译、竞态检测、静态检查、Mermaid 实际渲染、本地链接与 `git diff --check`。
- v3 快照与 `UNKNOWN` 前置能力的代码验证以固定持续集成任务为准；持续集成成功不得被解释为后备后继执行已经实现。

## 未实现与后续章节

- 无 `services/payment/internal/fallback`、`fallback_from_attempt_id` 唯一后继事务、正式 `RoutePlanDecision` 与 `FallbackExecutionBinding`，也没有跨支付服务商的 Compose 后备路由端到端测试。
- 当前代码尚无路由服务所有且只能否决的护栏复检与健康度子系统；未来检查只能允许或否决，不能新增、重排、重评分或延长计划，计划或证据过期、检查不可用时必须停止。
