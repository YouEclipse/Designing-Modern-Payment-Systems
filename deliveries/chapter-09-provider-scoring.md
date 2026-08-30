# 第 9 章《服务商评分与风险感知路由》交付记录

## 章节契约

- 核心命题：路由是固定阶段的受约束优化；上下文、准入筛选、规则与风险只能收缩或停止，成本与表现评分只能比较剩余的精确连接。
- 演进线：可互相补偿的加权和 → 风险与能力硬闸门 → 绑定主体的风险合同 → 完整特征批次 → 版本化整数目标 → 整批降级 → 稳定选择与执行期限。
- 本章纵向切片：`routing_context.v2`、共享 `risk_decision.v1`、路由服务 v3 约束流水线、成本与表现目标、严格的支付侧消费者与原子决策证据保存。
- 后续范围：动态健康度留第 13 章；独立决策存储投影器留第 14 章；后备后继事务留第 12 章；真实欺诈系统与制品控制平面、实验和机器学习不在本章实现。

## 配套实现

- 代码仓库：`YouEclipse/payments`。
- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- 持续集成：[成功](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 接口：支付服务显式选择 v3 写入方；路由服务 `/internal/v3/routing-decisions` 严格读取 `routing_context.v2`，任何错误都不回退到 v2 或 v1。

## 主张—实现—验证映射

工程文件相对于固定实现。

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 风险主体、动作、来源、版本、原因与失效时间构成不含凭证的共享合同 | `internal/riskdecision/decision.go`；支付侧 `internal/routingcontext/risk.go`；路由侧 `internal/routingcontext/context.go` | `TestRiskIdentityFreshnessAndReasonsFailClosed`；`TestRiskContextGoldenAndSubjectBinding` |
| `implemented` | 本地风险夹具只允许在本地与测试环境使用；缺少真实来源时拒绝执行 | `services/payment/cmd/api/routing.go`、`internal/routingcontext/risk.go` | `TestRoutingBuilderWiresExplicitVersionAndRiskBoundary`；`TestBuilderRiskIsExplicitAndCredentialFree` |
| `implemented` | `BLOCK`、`MANUAL_REVIEW` 和当前 `REQUIRE_3DS` 在查找特征前停止 | `internal/riskdecision/decision.go`；路由侧 `decision/engine.go` | `TestEveryNonAllowRiskActionStopsBeforeFeatureEvaluation`；支付错误映射测试 |
| `implemented` | 优化清单不能新增候选通道或让连接修订漂移 | `services/router/internal/decision/optimization.go` | `TestOptimizationManifestAndClassificationBindExactRevision` |
| `implemented` | 评分使用 `ConservativePPM + SucceededCostMicros` 的版本化整数目标 | `services/router/internal/scoring/types.go`、`score.go` | `TestNormalizationGoldenRoundingClampingAndExtremeRange`；`TestFullScoreHasExactIntegerGoldenAndAllProvenance` |
| `implemented` | 每个允许身份必须有完整且语义一致的成本与表现记录 | `services/router/internal/scoring/validate.go` | `TestMalformedEvidenceNeverGetsPermissiveFallback`；`TestDirectedFXEvidenceAndHierarchyArePreserved` |
| `implemented` | 特征不可用时整批采用 `static_order` 或 `reject`；畸形证据不能宽松降级 | `services/router/internal/scoring/profile.go`、`score.go` | `TestUnavailableFeatureIsAWholeBatchAction`；非法批次测试 |
| `implemented` | 稳定选择与输入顺序无关 | `services/router/internal/scoring/score.go` | `TestStableTieUsesRuleRankThenCandidateIDAndNeverInputOrder`；`TestConcurrentScoringIsDeterministicAndRaceFree` |
| `implemented` | 决策编号与 `executable_until` 绑定全部硬闸门、优化证据和最早失效时间 | `services/router/internal/decision/engine.go`、`optimization.go` | `TestOptimizationSnapshotDigestOrderOwnershipAndConcurrentReplay`；`TestExecutableAtRechecksTheWholeScoredBatchAfterEvaluation`；`TestExecutableAtRechecksFXClassificationAndPerformance` |
| `implemented` | 支付服务严格验证 v3 信封且不重算路由数学；从 JSONB 重读时以有类型的语义重建复验内容编号 | `services/payment/internal/clients/router.go`、`router_optimization.go` | `TestUniqueJSONRejectsAmbiguityAtEveryDepth`；`TestRouterV3RejectsIncompleteOrStateContradictoryOptimizationEvidence`；`TestVersionedDecisionContentIDUsesOptimizationSemanticsAcrossObjectReordering`；`TestVersionedDecisionContentIDRejectsAmbiguousOptimizationJSON` |
| `implemented` | 上下文、决策、风险、优化与候选顺序在首次执行前原子固定 | `services/payment/internal/postgres/routing_evidence.go`、迁移 `006_risk_routing_evidence.sql` | `TestRoutingEvidenceBindsLockedIntentAndSelectedExecutionIdentity`；`TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |
| `target_contract` | 后备路由只能针对固定路由计划刷新安全否决证据，不能重建候选集合或排序 | 本章决策的 `candidate_order` 合同；后续由支付服务拥有后备执行 | 本章正文不把候选顺序宣称为已实现的后继事务；第 12 章需另给“禁止重排”测试 |

## 图表与配图

- 3 个 Mermaid 覆盖硬约束与软偏好流水线、风险执行时序和当前 `REQUIRE_3DS` 的“失败即拒绝”边界，已用 Mermaid 命令行工具实际渲染。
- 本章没有新增装饰性位图；硬闸门与数据流由可解析 Mermaid 精确表达。
- 4 个 Go 教学代码块覆盖朴素加权和、风险闸门、整数评分和稳定选择；评分器字段已映射真实 `ConservativePPM`、`SucceededCostMicros`，风险示例与当前不支持 3DS 的行为一致。

## 发布验证

- 本章 4 个 Go 代码块由正文提取为 4 个独立包，编译、竞态检测与静态检查已通过；附加测试覆盖风险证据不足即拒绝与生产字段映射。
- 本地 Markdown 链接、Mermaid、支付术语和固定实现链接已检查。
- 固定实现的单元测试、竞态检测、静态检查、OpenAPI、PostgreSQL 集成测试与 Compose 端到端测试，已由持续集成任务 `33041147119` 在固定提交上验证成功。

## 能力边界

当前 `DevelopmentRisk` 不是欺诈引擎，费用与表现是合成夹具，支付服务商由模拟器扮演。系统没有 3DS 续程、手动请款、动态健康度、生产制品存储、签名、审批、热发布、实验、机器学习、真实 PSP 沙箱或生产支付能力。支付服务已经保存完整决策快照，但独立可查询的决策存储、发件箱投影器与后备后继事务尚未实现。

迁移 `006_risk_routing_evidence.sql` 会在失败时拒绝继续：若迁移 004 时代合法的旧快照缺少 `candidate_order`，或第一项不等于已固定的选中连接，整个升级事务会因约束错误回滚，不能用单元素数组猜测历史事实。生产升级必须先按[支付服务 006 路由证据升级运行手册](https://github.com/YouEclipse/payments/blob/bc4afde24da708d22f67da72a90a1557b1477173/docs/runbooks/payment-006-risk-routing-evidence.md)做只读预检、可恢复备份和隔离恢复演练；只有可追溯的原始路由服务响应才能用于受控修复。
