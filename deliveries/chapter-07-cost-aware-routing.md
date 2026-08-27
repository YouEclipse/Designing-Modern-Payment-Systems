# 第 7 章《Cost-Aware Routing》交付记录

## 章节契约

- 核心命题：Expected Cost 是 Router 对已经 eligible 的精确 Candidate 所做的版本化、可分解估计；它不是资金状态、账本、Settlement 或 Actual Cost。
- 演进线：单一比例费 → fixed fee 价格交叉 → Candidate/revision/operation 合同作用域 → 跨境与定向 FX → attempt/success/failure 条件情景 → 整批降级与执行期限。
- 本章纵向切片：`fee_catalog.v1` compiler、定点估价器、费用证据进入 Router v3 Optimization，以及 Scorer 对成功情景的显式选择。
- 后续范围：真实费率摄取、审批发布与阶梯价留第 20 章；Actual Cost、账单与 Settlement 归因留第 18 章；Fallback 链成本不是本章 Scorer 的隐含输入。

## 配套实现

- Repository：`YouEclipse/payments`。
- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[success](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- API：Router `/internal/v3/routing-decisions` 在硬闸门之后返回 Cost/Performance/Scoring 的完整 Optimization evidence；Payment 严格验证后将完整语义固定为 JSONB，重读时以 typed canonical reconstruction 验证 content ID，而不把原始字节或对象 key 顺序当作权威。

## 主张—实现—验证映射

工程文件相对于固定实现。

| status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 费率绑定精确 Candidate revision、operation、币种与半开有效区间 | `services/router/internal/cost/types.go`、`artifact.go` | `TestCostCompileRejectsAmbiguousOrUnboundArtifacts`；`TestScheduleHalfOpenIntervalsDoNotDependOnLoadingOrder` |
| `implemented` | 金额、BPS、舍入与 FX 全部使用定点/有理数；溢出不返回部分估值 | `services/router/internal/cost/arithmetic.go`、`estimate.go` | `TestFixedPointRoundingAndRepresentableLargeQuotients`；`TestMulDivAgainstExactRationalProperties`；`FuzzFixedPointDoesNotWrap` |
| `implemented` | Attempt、Success-only、Failure-only 费用互斥保存；成功与失败费用不相加 | `services/router/internal/cost/types.go`、`estimate.go` | `TestFeeScenariosReproducePriceCrossingWithoutInventingProbability` |
| `implemented` | 缺 Schedule、陈旧版本、未知跨境与缺 FX 不会变成零成本 | `services/router/internal/cost/estimate.go` | `TestCostUnknownAndOverflowNeverExposePartialZeroTotals`；`TestUnknownCrossBorderNeverBecomesZeroCostDomestic` |
| `implemented` | 跨境分类保留 country field、rule/source/revision 与 expiry | `services/router/internal/decision/optimization.go` | `TestOptimizationManifestAndClassificationBindExactRevision`；classification expiry tests |
| `implemented` | Scorer 明确消费 `SucceededCostMicros`，同时保留完整原始 Breakdown | `services/router/internal/scoring/types.go`、`score.go` | `TestFullScoreHasExactIntegerGoldenAndAllProvenance`；`TestFeatureEvidenceAndCompiledProfileAreOwnedSnapshots` |
| `implemented` | 任一合法 Feature 不可用只能整批 `static_order` 或 `reject` | `services/router/internal/scoring/score.go`、`validate.go` | `TestUnavailableFeatureIsAWholeBatchAction`；malformed batch rejection tests |
| `implemented` | 费用证据的最早失效时间限制首次 dispatch 权威 | `services/router/internal/decision/optimization.go`、`engine.go` | `TestExecutableAtRechecksTheWholeScoredBatchAfterEvaluation`；FX/classification expiry tests |
| `implemented` | Payment 不重算 Router 数学；严格验证并以 JSONB 语义完整固定 v3 envelope，重读时按 typed wire 重建 content ID | `services/payment/internal/clients/router_optimization.go`、`internal/postgres/routing_evidence.go` | `TestRouterV3ConsumesSharedRouterSuccessGoldenWithoutLoss`；`TestVersionedDecisionContentIDUsesOptimizationSemanticsAcrossObjectReordering`；`TestVersionedDecisionContentIDRejectsAmbiguousOptimizationJSON`；`TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |

## 图表与配图

- 3 个 Mermaid 覆盖 Eligibility→Cost、估价数据流和 Expected/Actual 双证据链，已用 Mermaid CLI 实际渲染。
- Ian 正文配图 `assets/cost-aware-routing-illustrations/01-hidden-receipts.png` 为 1672×941，表达“低标价后仍拖着 fixed/cross-border/FX 小票”；它不替代费用契约图。
- 3 个 Go 教学块分别覆盖比例费反例、同币种定点估价和完整集合排序；正文明确映射真实 `ConnectionRevision`、三种费用情景与 `SucceededCostMicros`。

## 发布验证

- 本章 3 个 Go 块由正文提取为 3 个独立 package，compile、race 与 vet 已通过。
- 本地 Markdown 链接、图片路径、Mermaid、支付术语和固定实现链接已检查。
- 固定实现的 unit、race、vet、OpenAPI、PostgreSQL integration 与 Compose E2E 已由 CI run `33041147119` 在固定 SHA 上验证为 success。

## 能力边界

仓库中的 Fee/FX Artifact 是合成 fixture，不来自真实 Provider 合同、实时 FX 或审批控制面。`SucceededCostMicros` 是成功情景 Expected Cost，不是 Invoice、Settlement、利润或成功支付全链路成本。Cost package 可表达 `authorize` Schedule，不代表 manual Capture 已实现；当前 v3 资金切片仍只有 automatic-capture `SALE`。系统没有阶梯返利、Actual Cost Attribution、真实 PSP sandbox 或 live payment 能力。
