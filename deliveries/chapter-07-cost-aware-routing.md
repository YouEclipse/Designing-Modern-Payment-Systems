# 第 7 章《成本感知路由》交付记录

## 章节契约

- 核心命题：预期成本是路由服务对已经通过准入筛选的精确候选通道所做的版本化、可分解估计；它不是资金状态、账本、结算或实际成本。
- 演进线：单一比例费 → 固定费造成价格交叉 → 候选通道、修订与资金操作的合同作用域 → 跨境与定向外汇转换 → 尝试、成功与失败的条件情景 → 整批降级与执行期限。
- 本章纵向切片：`fee_catalog.v1` 编译器、定点估价器、进入路由服务 v3 优化证据的费用，以及评分器对成功情景的显式选择。
- 后续范围：真实费率摄取、审批发布与阶梯价留第 20 章；实际成本、账单与结算归因留第 18 章；后备路由链成本不是本章评分器的隐含输入。

## 配套实现

- 代码仓库：`YouEclipse/payments`。
- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- 持续集成：[成功](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 接口：路由服务 `/internal/v3/routing-decisions` 在硬闸门之后返回成本、表现与评分的完整优化证据；支付服务严格验证后将完整语义固定为 JSONB，重读时以有类型的规范重建来验证内容编号，而不把原始字节或对象键顺序当作权威。

## 主张—实现—验证映射

工程文件相对于固定实现。

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 费率绑定精确候选通道修订、资金操作、币种与半开有效区间 | `services/router/internal/cost/types.go`、`artifact.go` | `TestCostCompileRejectsAmbiguousOrUnboundArtifacts`；`TestScheduleHalfOpenIntervalsDoNotDependOnLoadingOrder` |
| `implemented` | 金额、基点、舍入与外汇转换全部使用定点数或有理数；溢出时不返回部分估值 | `services/router/internal/cost/arithmetic.go`、`estimate.go` | `TestFixedPointRoundingAndRepresentableLargeQuotients`；`TestMulDivAgainstExactRationalProperties`；`FuzzFixedPointDoesNotWrap` |
| `implemented` | 尝试、仅成功、仅失败费用互斥保存；成功与失败费用不相加 | `services/router/internal/cost/types.go`、`estimate.go` | `TestFeeScenariosReproducePriceCrossingWithoutInventingProbability` |
| `implemented` | 费率表缺失、版本陈旧、跨境属性未知或缺少外汇汇率时不会变成零成本 | `services/router/internal/cost/estimate.go` | `TestCostUnknownAndOverflowNeverExposePartialZeroTotals`；`TestUnknownCrossBorderNeverBecomesZeroCostDomestic` |
| `implemented` | 跨境分类保留国家字段、规则、来源、修订与失效时间 | `services/router/internal/decision/optimization.go` | `TestOptimizationManifestAndClassificationBindExactRevision`；分类失效时间测试 |
| `implemented` | 评分器明确消费 `SucceededCostMicros`，同时保留完整原始分解 | `services/router/internal/scoring/types.go`、`score.go` | `TestFullScoreHasExactIntegerGoldenAndAllProvenance`；`TestFeatureEvidenceAndCompiledProfileAreOwnedSnapshots` |
| `implemented` | 任一合法特征不可用时，只能整批采用 `static_order` 或 `reject` | `services/router/internal/scoring/score.go`、`validate.go` | `TestUnavailableFeatureIsAWholeBatchAction`；畸形批次拒绝测试 |
| `implemented` | 费用证据的最早失效时间限制首次派发权威 | `services/router/internal/decision/optimization.go`、`engine.go` | `TestExecutableAtRechecksTheWholeScoredBatchAfterEvaluation`；外汇与分类失效时间测试 |
| `implemented` | 支付服务不重算路由数学；严格验证并以 JSONB 语义完整固定 v3 信封，重读时按有类型的线格式重建内容编号 | `services/payment/internal/clients/router_optimization.go`、`internal/postgres/routing_evidence.go` | `TestRouterV3ConsumesSharedRouterSuccessGoldenWithoutLoss`；`TestVersionedDecisionContentIDUsesOptimizationSemanticsAcrossObjectReordering`；`TestVersionedDecisionContentIDRejectsAmbiguousOptimizationJSON`；`TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |

## 图表与配图

- 3 个 Mermaid 覆盖准入筛选→成本、估价数据流和预期成本与实际成本的双证据链，已用 Mermaid 命令行工具实际渲染。
- Ian 正文配图 `assets/cost-aware-routing-illustrations/01-hidden-receipts.png` 为 1672×941，表达“低标价后仍拖着固定费、跨境费与外汇转换小票”；它不替代费用契约图。
- 3 个 Go 教学代码块分别覆盖比例费反例、同币种定点估价和完整集合排序；正文明确映射真实 `ConnectionRevision`、三种费用情景与 `SucceededCostMicros`。

## 发布验证

- 本章 3 个 Go 代码块由正文提取为 3 个独立包，编译、竞态检测与静态检查已通过。
- 本地 Markdown 链接、图片路径、Mermaid、支付术语和固定实现链接已检查。
- 固定实现的单元测试、竞态检测、静态检查、OpenAPI、PostgreSQL 集成测试与 Compose 端到端测试，已由持续集成任务 `33041147119` 在固定提交上验证成功。

## 能力边界

仓库中的费用与外汇制品是合成夹具，不来自真实支付服务商合同、实时外汇数据或审批控制面。`SucceededCostMicros` 是成功情景下的预期成本，不是账单、结算、利润或成功支付全链路成本。成本包可以表达 `authorize` 费率表，不代表手动请款已经实现；当前 v3 资金切片仍只有自动请款 `SALE`。系统没有阶梯返利、实际成本归因、真实 PSP 沙箱或生产支付能力。
