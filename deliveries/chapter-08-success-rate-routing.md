# 第 8 章《Success-Rate Routing》交付记录

## 章节契约

- 核心命题：线上只能消费带身份、分段、样本、pending coverage、窗口、不确定性和版本的条件表现证据；历史路由下的观测不能被包装成 Candidate 的因果优势。
- 演进线：Provider 总榜 → 精确 connection identity → outcome/denominator 合同 → 固定 hierarchy → per-level threshold、pending coverage、shrinkage/Wilson → stale/prior 降级。
- 本章纵向切片：严格编译 `performance.v1`，读取 fixture-backed observational snapshot，并把 typed evidence 送入 Router v3 Scoring batch。
- 后续范围：真实 outcome producer 与指标流水线留第 15、17 章；因果实验和反事实支持留第 16、21 章；分钟级动态 Health 留第 13 章。

## 配套实现

- Repository：`YouEclipse/payments`。
- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[success](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 契约：Segment 使用 `candidate_id + connection_revision + country + payment_method_context.network + currency`，并显式声明 `country_field`、SALE/Authorization cohort 与 denominator version。

## 主张—实现—验证映射

工程文件相对于固定实现。

| status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | Candidate ID 与 connection revision 是不可拆统计身份 | `services/router/internal/performance/types.go`、`artifact.go` | `TestPerformanceCohortAndRevisionCannotSilentlyMix`；`TestPerformanceCompileRejectsAmbiguousStatisticsAndContracts` |
| `implemented` | `network` 来自 `payment_method_context.network`，country 语义由 `country_field` 固定 | `services/router/internal/performance/types.go`；`decision/optimization.go` | `TestUnknownGeographyAndNetworkNeverBecomeKnownEmptyKeys`；`TestOptimizationManifestAndClassificationBindExactRevision` |
| `implemented` | 固定 L0→L4 hierarchy 不挑成功率最高的 sibling/parent | `services/router/internal/performance/lookup.go` | `TestFixedHierarchyRejectsOneOfOneAndDoesNotChooseBestParent`；`TestColdStartPriorIsNotFabricatedCandidateHistory` |
| `implemented` | 每层独立 sample threshold；pending 不先算失败，而是降低 coverage | `services/router/internal/performance/artifact.go`、`lookup.go`、`statistics.go` | `TestPendingOutcomesDegradeCoverageWithoutBecomingDeclines` |
| `implemented` | Window、watermark、publication age 与 data-window age 都必须新鲜 | `services/router/internal/performance/artifact.go`、`lookup.go` | `TestFreshPublicationCannotHideAnOldWindow` |
| `implemented` | Shrinkage 借用同一固定链的祖先；Wilson 区间不伪装成 posterior | `services/router/internal/performance/statistics.go` | `TestShrinkageDoesNotTurnOneSuccessIntoCertainSuperiority`；`TestObservedWilsonIntervalIsNotAShrinkagePosterior`；`FuzzStatisticalEvidenceRemainsBounded` |
| `implemented` | Scorer 消费保守 `ConservativePPM`，但保留概率、区间、样本与 observational-only 解释 | `services/router/internal/performance/types.go`；`scoring/score.go` | `TestFullScoreHasExactIntegerGoldenAndAllProvenance`；`TestDirectedFXEvidenceAndHierarchyArePreserved` |
| `implemented` | 不合格 Candidate 永不进入 Performance lookup，Feature 不能复活硬排除 | `services/router/internal/decision/engine.go`、`optimization.go` | `TestOptimizationCannotReviveAnEligibilityOrRuleExclusion` |
| `implemented` | Artifact immutable、canonical、可并发重放 | `services/router/internal/performance/artifact.go` | `TestPerformanceSnapshotCanonicalOwnershipAndConcurrentReplay` |

## 图表与配图

- 4 个 Mermaid 覆盖选择偏差闭环、观测生产链、固定 hierarchy 与 Eligibility-before-Performance，已用 Mermaid CLI 实际渲染。
- Ian 正文配图 `assets/success-rate-routing-illustrations/01-one-ranking-many-tracks.png` 为 1672×941，用于建立“总榜不是因果”的直觉。
- 2 个 Go 教学块覆盖固定 lookup 与 shrinkage/Wilson；正文明确说明教学代码的 known-only/no-prior 简化，并映射 production revision、state、threshold、pending、country/network、window 与 conservative PPM。

## 发布验证

- 本章 2 个 Go 块由正文提取为 1 个独立 package，compile、race 与 vet 已通过；附加测试覆盖 revision、pending coverage 与双年龄边界。
- 本地 Markdown 链接、图片路径、Mermaid、支付术语和固定实现链接已检查。
- 固定实现的 unit、race、vet、OpenAPI、PostgreSQL integration 与 Compose E2E 已由 CI run `33041147119` 在固定 SHA 上验证为 success。

## 能力边界

当前 Snapshot、样本和 Outcome 都是合成 fixture。Performance package 可校验 `authorize + manual` 统计 schema，但当前 v3 只安装 `sale + automatic` Snapshot，manual Capture 资金链并未实现。代码没有真实 PSP 结果归一化与 logical-operation dedup producer，没有流处理、Feature Store、Artifact 签名/审批/发布、动态 Health、实验或 OPE。它只能证明 observational reader contract，不能证明真实 Authorization uplift、因果收益、Provider benchmark 或 live payment 能力。
