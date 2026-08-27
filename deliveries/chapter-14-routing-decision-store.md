# 第 14 章《Routing Decision Store》交付记录

## 章节契约

- 核心命题：历史选择必须由决策时封存的不可变事实回答；Payment 本地 execution snapshot 是执行权威，远端 Store 只能异步 append/query。
- 演进线：`selected_provider` 日志 → 完整 v3 execution snapshot → canonical source bytes/full hash → outbox → append-only projection/quarantine。
- 当前代码增量：完整 Context/Decision/Risk/Optimization/Candidate order/`executable_until` 与 Attempt/Operation 原子固定并不可改写。
- 后续范围：canonical bytes/full SHA protocol、decision-specific outbox、Decision Store/projector，以及 `RoutePlanDecision`/Payment-authored `FallbackExecutionBinding` successor transaction 尚未实现。

## 配套实现

- Repository：`YouEclipse/payments`
- Implementation commit：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- CI run：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 纵向切片：Payment 在首次资金 I/O 前事务化固定完整 Router v3 JSON evidence 与 deadline；JSONB snapshot 是本地执行真相，但不是最终 canonical archive。

## 主张—实现—验证映射

| status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | Router/Payment 对同一 v3 Context/Decision golden 无损解析 | contract fixtures；Router `internal/httpapi/v2.go`；Payment `clients/router_v3_test.go` | `TestV3HTTPSuccessMatchesSharedContractGolden`、`TestRouterV3ConsumesSharedRouterSuccessGoldenWithoutLoss` |
| `implemented` | Context subject/digest、Risk ALLOW、Optimization order、selected revision 与 deadline 在事务内复验 | `services/payment/internal/postgres/routing_evidence.go` | `TestRoutingEvidenceBindsLockedIntentAndSelectedExecutionIdentity`、`TestRouterV3RejectsTamperedOrCredentialLikeEvidence` |
| `implemented` | 完整 Context/Decision snapshots、Attempt/Operation 与 deadline 原子固定，之后不可改写 | migration `006_risk_routing_evidence.sql`；`services/payment/internal/postgres/store.go` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |
| `implemented` | 首次资金 I/O 在 `executable_until` 边界后 fail closed；已 dispatch recovery 不被错误重路由 | Payment recovery/store | `TestRoutingEvidenceExpiresBeforeFirstProviderDispatch`、`TestExpiredRoutingAuthorityDoesNotBlockAlreadyDispatchedSameKeyRecovery` |
| `implemented` | Payment 在 JSONB 重读后按 typed Optimization wire 语义重建 `rtd_` material：对象 key 重排仍通过，字段值或有序数组变化失败，ambiguous JSON 被拒绝；该 ID 仍只是截断关联 ID | `services/payment/internal/clients/router_optimization.go`；Router Decision encoder | `TestVersionedDecisionContentIDUsesOptimizationSemanticsAcrossObjectReordering`、`TestVersionedDecisionContentIDRejectsAmbiguousOptimizationJSON`；shared golden 与 tamper tests |
| `teaching_model` | canonical Envelope 用 `Outcome` 判别联合区分 selected 与六类负向结果；负向结果不伪造 Attempt/Plan/Selection，可保存 `not_exposed` assignment | 本章 Go blocks | chapter14 `envelope_test.go` 的 outcome branch、assignment-only 与 contradiction tests |
| `teaching_model` | pinning 与 projection 的教学合同可编译/可反例验证 | 本章 Go blocks | chapter14 `pinning_test.go`、`projection_test.go` |
| `implemented` | 双 captured 保留业务成功，独立 integrity review，不覆盖 Router Plan 或 Payment fallback binding | Payment domain model | `TestConflictingTerminalOutcomeRequiresReview`、`TestReviewRequiredIsNotIgnoredAfterTerminalSuccess` |

## 图表与配图

- 4 个 Mermaid：历史/反事实差异、本地 snapshot+outbox、摄取冲突、Fallback Decision lineage。
- Ian 正文配图：`assets/routing-decision-store-illustrations/01-the-decision-black-box.png`，1672×941；表达“箱内历史不随当前配置变化”。
- 4 个 Go 块由正文提取，chapter14 tests 覆盖 selected/负向 outcome 判别联合、assignment-only、canonical envelope、transaction pinning 与 projection ingest。

## 发布验证

- 书稿侧执行章节 Go compile/race/vet、Mermaid 实际渲染、本地链接、图片路径/尺寸与 `git diff --check`。
- v3 local execution snapshot 的 unit/integration/Compose 证据以固定 CI run 为准；CI success 不得被解释为远端 Decision Store 已实现。

## 未实现与后续章节

- JSONB 不保存原始 canonical bytes；当前 typed semantic reconstruction 只用于 Router/Payment 的现有 content-ID binding，`rtd_` 仍只有截断 ID，没有 raw canonical archive、持久化 full SHA、canonicalization version/source row 或跨语言 canonicalization protocol。
- 无 `routing_decision.recorded` outbox、append-only source/projection/quarantine store 或 projector/rebuild。
- 无专用 Fallback successor transaction；未来 Payment-authored `FallbackExecutionBinding` 只能绑定 Router-authored `RoutePlanDecision` 的固定下一项和新鲜 Risk/guardrail 证据。Router-owned revalidation 只能 veto，不能产生新 Candidate Set/Ranking、重排或续期；过期/不可用必须 Stop。
