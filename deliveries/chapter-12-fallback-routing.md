# 第 12 章《Fallback Routing》交付记录

## 章节契约

- 核心命题：只有旧 Attempt `terminal + side_effect=none + allowlisted reason` 且所有 guardrail 通过，Payment 才可沿固定 Route Plan 创建一个新 Attempt。
- 演进线：遇错换 PSP → certainty/reason 双闸门 → 固定 Candidate order → 新身份 successor → 并发唯一 transaction。
- 当前代码基线：v3 完整 Context/Decision/Risk/Optimization/order/deadline 原子快照与 UNKNOWN 默认阻断。
- 明确设计项：Router-authored `RoutePlanDecision`、Router-owned guardrail-only revalidation、Payment-authored `FallbackExecutionBinding`、successor transaction 与跨 Provider E2E 尚未实现。

## 配套实现

- Repository：`YouEclipse/payments`
- Implementation commit：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- CI run：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 本章纵向切片：本轮只交付 Fallback 的安全前置条件与可编译教学模型，不宣称代码仓库已经能创建 Attempt 2。

## 主张—实现—验证映射

| status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | UNKNOWN/possible 只恢复原 Operation，默认禁止新 Attempt | `services/payment/internal/domain/model.go`；`services/payment/internal/recovery/planner.go` | `TestUnknownBlocksNewConfirmationAndConverges`、`TestRecoveryActionTable` |
| `implemented` | v3 Context、Risk、Optimization、Candidate order 与 deadline 在首次 Attempt 事务中原子固定 | `services/payment/internal/postgres/routing_evidence.go`；migration `006_risk_routing_evidence.sql`；`store.go` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten`、`TestRoutingEvidenceBindsLockedIntentAndSelectedExecutionIdentity` |
| `implemented` | Risk block/review/require-3DS 不能产生 executable Candidate | Router v3 与 Payment strict client | `TestRouterV3RiskActionsCannotCreateAnExecutableCandidate`、`TestV3HTTPSuccessMatchesSharedContractGolden` |
| `teaching_model` | Fallback 教学 Planner 只消费 Router-authored `RoutePlanDecision` 的固定下一项；revalidation 只能 veto，不能重排或续期 | 本章 `package fallback` Go block | `scripts/go-examples/tests/chapter12/fallback_test.go` |
| `target_contract` | 获准 successor 由 Payment 写 `FallbackExecutionBinding`，引用原 Plan、前序 Attempt、固定 next 与新鲜 guardrail 证据，不冒充 Router Decision | 本章 authority contract；未来 Payment successor transaction | 正文 authority/expiry/veto Failure Modes；当前无 runtime test |
| `teaching_model` | successor 必须使用新 Attempt/Operation/key，不能改写旧身份 | 本章 `package successor` Go block | `scripts/go-examples/tests/chapter12/successor_test.go` |
| `implemented` | 双 confirmed 不覆盖 captured 事实，业务成功与 integrity review 分轴 | `services/payment/internal/domain/model.go` | `TestConflictingTerminalOutcomeRequiresReview`、`TestManualReviewStopsAutomaticRecoveryAndPreservesFunds` |

## 图表与配图

- 4 个 Mermaid：误切换双扣、Attempt chain、Planner gate、并发 successor transaction。
- 本章未添加 Ian 插图；资金路径与门禁更适合精确流程图。
- 3 个 Go 块由正文提取，chapter12 tests 覆盖 naive 反例、Planner 与 successor identity。

## 发布验证

- 书稿侧执行章节 Go compile/race/vet、Mermaid 实际渲染、本地链接与 `git diff --check`。
- v3 snapshot/UNKNOWN 前置能力的代码验证以固定 CI run 为准；CI success 不得被解释为 Fallback successor 已实现。

## 未实现与后续章节

- 无 `services/payment/internal/fallback`、`fallback_from_attempt_id` 唯一 successor transaction、正式 `RoutePlanDecision`/`FallbackExecutionBinding` 或跨 Provider Fallback Compose E2E。
- 当前代码尚无 Router-owned guardrail-only revalidation 与 Health subsystem；未来检查只能 allow/veto，不能新增、重排、重评分或延长 Plan，Plan/证据过期或检查不可用必须 Stop。
