# 第 9 章《Provider Scoring 与 Risk-Aware Routing》交付记录

## 章节契约

- 核心命题：路由是固定阶段的受约束优化；Context、Eligibility、Rule 与 Risk 只能收缩或停止，Cost/Performance Scoring 只能比较剩余精确连接。
- 演进线：可补偿 weighted sum → Risk/Capability hard gate → subject-bound Risk contract → 完整 Feature batch → versioned integer Objective → whole-batch degradation → stable selection 与执行期限。
- 本章纵向切片：`routing_context.v2`、共享 `risk_decision.v1`、Router v3 constrained pipeline、Cost/Performance Objective、严格 Payment consumer 与原子决策证据保存。
- 后续范围：动态 Health 留第 13 章；独立 Decision Store projector 留第 14 章；Fallback successor transaction 留第 12 章；真实 Fraud/Artifact control plane、实验和 ML 分别不在本章实现。

## 配套实现

- Repository：`YouEclipse/payments`。
- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[success](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- API：Payment 显式选择 v3 writer；Router `/internal/v3/routing-decisions` 严格读取 `routing_context.v2`，任何错误都不回退 v2/v1。

## 主张—实现—验证映射

工程文件相对于固定实现。

| status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | Risk subject、Action、source、version、reason 与 expiry 是 credential-free 共享合同 | `internal/riskdecision/decision.go`；Payment `internal/routingcontext/risk.go`；Router `internal/routingcontext/context.go` | `TestRiskIdentityFreshnessAndReasonsFailClosed`；`TestRiskContextGoldenAndSubjectBinding` |
| `implemented` | local Risk fixture 只允许 local/test，缺真实来源 fail closed | `services/payment/cmd/api/routing.go`、`internal/routingcontext/risk.go` | `TestRoutingBuilderWiresExplicitVersionAndRiskBoundary`；`TestBuilderRiskIsExplicitAndCredentialFree` |
| `implemented` | `BLOCK`、`MANUAL_REVIEW`、当前 `REQUIRE_3DS` 在 Feature lookup 前停止 | `internal/riskdecision/decision.go`；Router `decision/engine.go` | `TestEveryNonAllowRiskActionStopsBeforeFeatureEvaluation`；Payment error mapping tests |
| `implemented` | Optimization manifest 不能新增 Candidate 或漂移 connection revision | `services/router/internal/decision/optimization.go` | `TestOptimizationManifestAndClassificationBindExactRevision` |
| `implemented` | Scoring 使用 `ConservativePPM + SucceededCostMicros` 的版本化整数 Objective | `services/router/internal/scoring/types.go`、`score.go` | `TestNormalizationGoldenRoundingClampingAndExtremeRange`；`TestFullScoreHasExactIntegerGoldenAndAllProvenance` |
| `implemented` | 每个 allowed identity 必须有完整、同一语义的 Cost/Performance row | `services/router/internal/scoring/validate.go` | `TestMalformedEvidenceNeverGetsPermissiveFallback`；`TestDirectedFXEvidenceAndHierarchyArePreserved` |
| `implemented` | unavailable 是整批 `static_order` 或 `reject`；malformed 不能 permissive fallback | `services/router/internal/scoring/profile.go`、`score.go` | `TestUnavailableFeatureIsAWholeBatchAction`；invalid batch tests |
| `implemented` | Stable selection 与输入顺序无关 | `services/router/internal/scoring/score.go` | `TestStableTieUsesRuleRankThenCandidateIDAndNeverInputOrder`；`TestConcurrentScoringIsDeterministicAndRaceFree` |
| `implemented` | Decision ID 与 `executable_until` 绑定全部硬闸门、Optimization 和最早失效证据 | `services/router/internal/decision/engine.go`、`optimization.go` | `TestOptimizationSnapshotDigestOrderOwnershipAndConcurrentReplay`；`TestExecutableAtRechecksTheWholeScoredBatchAfterEvaluation`；`TestExecutableAtRechecksFXClassificationAndPerformance` |
| `implemented` | Payment 严格验证 v3 envelope 且不重算 Router 数学；JSONB 重读以 typed semantic reconstruction 复验 content ID | `services/payment/internal/clients/router.go`、`router_optimization.go` | `TestUniqueJSONRejectsAmbiguityAtEveryDepth`；`TestRouterV3RejectsIncompleteOrStateContradictoryOptimizationEvidence`；`TestVersionedDecisionContentIDUsesOptimizationSemanticsAcrossObjectReordering`；`TestVersionedDecisionContentIDRejectsAmbiguousOptimizationJSON` |
| `implemented` | Context/Decision/Risk/Optimization/candidate order 在首次执行前原子固定 | `services/payment/internal/postgres/routing_evidence.go`、迁移 `006_risk_routing_evidence.sql` | `TestRoutingEvidenceBindsLockedIntentAndSelectedExecutionIdentity`；`TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |
| `target_contract` | Fallback 只能对固定 Route Plan 刷新安全 veto，不能重新建立 Candidate set/ranking | 本章 Decision `candidate_order` 合同；后续 Payment fallback owner | 本章正文不把 Candidate order 宣称为已实现的 successor transaction；第 12 章需另给 no-rerank 测试 |

## 图表与配图

- 3 个 Mermaid 覆盖 hard/soft pipeline、Risk 执行时序和当前 `REQUIRE_3DS` fail-closed 边界，已用 Mermaid CLI 实际渲染。
- 本章没有新增装饰性位图；硬闸门与数据流由可解析 Mermaid 精确表达。
- 4 个 Go 教学块覆盖 naive weighted sum、Risk gate、整数 scoring 和 stable selection；Scorer 字段已映射真实 `ConservativePPM`、`SucceededCostMicros`，Risk 示例与当前 3DS unsupported 行为一致。

## 发布验证

- 本章 4 个 Go 块由正文提取为 4 个独立 package，compile、race 与 vet 已通过；附加测试覆盖 Risk fail-closed 与 production 字段映射。
- 本地 Markdown 链接、Mermaid、支付术语和固定实现链接已检查。
- 固定实现的 unit、race、vet、OpenAPI、PostgreSQL integration 与 Compose E2E 已由 CI run `33041147119` 在固定 SHA 上验证为 success。

## 能力边界

当前 `DevelopmentRisk` 不是 Fraud Engine，Fee/Performance 是合成 fixture，Provider 是 Simulator。系统没有 3DS continuation、manual Capture、动态 Health、生产 Artifact Store/签名/审批/热发布、实验、ML、真实 PSP sandbox 或 live payment 能力。Payment 已保存完整决策快照，但独立可查询 Decision Store/outbox projector 与 Fallback successor transaction 尚未实现。

Migration `006_risk_routing_evidence.sql` 会 fail closed：若 migration 004 时代合法的旧 snapshot 缺少 `candidate_order`，或第一项不等于已固定的 selected connection，整个升级事务以约束错误回滚，不能用单元素数组猜测历史事实。生产升级必须先按 [Payment 006 路由证据升级 Runbook](https://github.com/YouEclipse/payments/blob/bc4afde24da708d22f67da72a90a1557b1477173/docs/runbooks/payment-006-risk-routing-evidence.md) 做只读预检、可恢复备份和隔离恢复演练；只有可追溯的原始 Router response 才能用于受控修复。
