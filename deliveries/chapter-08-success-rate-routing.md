# 第 8 章《成功率路由》交付记录

## 章节契约

- 核心命题：线上只能消费带身份、分段、样本、待定结果覆盖率、窗口、不确定性和版本的条件表现证据；历史路由下的观测不能被包装成候选通道的因果优势。
- 演进线：支付服务商总榜 → 精确连接身份 → 结果与分母合同 → 固定层级 → 分层阈值、待定结果覆盖率、收缩估计与 Wilson 区间 → 过期或先验降级。
- 本章纵向切片：严格编译 `performance.v1`，读取基于夹具的观察性快照，并把有类型的证据送入路由服务 v3 评分批次。
- 后续范围：真实结果生产器与指标流水线留第 15、17 章；因果实验和反事实支持留第 16、21 章；分钟级动态健康度留第 13 章。

## 配套实现

- 代码仓库：`YouEclipse/payments`。
- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- 持续集成：[成功](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 契约：分段使用 `candidate_id + connection_revision + country + payment_method_context.network + currency`，并显式声明 `country_field`、`SALE` 队列或授权队列，以及分母版本。

## 主张—实现—验证映射

工程文件相对于固定实现。

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 候选通道编号与连接修订是不可拆的统计身份 | `services/router/internal/performance/types.go`、`artifact.go` | `TestPerformanceCohortAndRevisionCannotSilentlyMix`；`TestPerformanceCompileRejectsAmbiguousStatisticsAndContracts` |
| `implemented` | `network` 来自 `payment_method_context.network`，国家字段的语义由 `country_field` 固定 | `services/router/internal/performance/types.go`；`decision/optimization.go` | `TestUnknownGeographyAndNetworkNeverBecomeKnownEmptyKeys`；`TestOptimizationManifestAndClassificationBindExactRevision` |
| `implemented` | 固定 L0→L4 层级不挑成功率最高的同级或父级分段 | `services/router/internal/performance/lookup.go` | `TestFixedHierarchyRejectsOneOfOneAndDoesNotChooseBestParent`；`TestColdStartPriorIsNotFabricatedCandidateHistory` |
| `implemented` | 每层有独立样本阈值；待定结果不预先算失败，而是降低覆盖率 | `services/router/internal/performance/artifact.go`、`lookup.go`、`statistics.go` | `TestPendingOutcomesDegradeCoverageWithoutBecomingDeclines` |
| `implemented` | 窗口、水位线、发布时间年龄与数据窗口年龄都必须足够新鲜 | `services/router/internal/performance/artifact.go`、`lookup.go` | `TestFreshPublicationCannotHideAnOldWindow` |
| `implemented` | 收缩估计借用同一固定链的祖先；Wilson 区间不伪装成后验分布 | `services/router/internal/performance/statistics.go` | `TestShrinkageDoesNotTurnOneSuccessIntoCertainSuperiority`；`TestObservedWilsonIntervalIsNotAShrinkagePosterior`；`FuzzStatisticalEvidenceRemainsBounded` |
| `implemented` | 评分器消费保守的 `ConservativePPM`，但保留概率、区间、样本与“仅为观察”的解释 | `services/router/internal/performance/types.go`；`scoring/score.go` | `TestFullScoreHasExactIntegerGoldenAndAllProvenance`；`TestDirectedFXEvidenceAndHierarchyArePreserved` |
| `implemented` | 不合格候选通道永不进入表现查找，特征不能复活硬排除 | `services/router/internal/decision/engine.go`、`optimization.go` | `TestOptimizationCannotReviveAnEligibilityOrRuleExclusion` |
| `implemented` | 制品不可变、规范化且可并发重放 | `services/router/internal/performance/artifact.go` | `TestPerformanceSnapshotCanonicalOwnershipAndConcurrentReplay` |

## 图表与配图

- 4 个 Mermaid 覆盖选择偏差闭环、观测生产链、固定层级与“先准入、后比较表现”，已用 Mermaid 命令行工具实际渲染。
- Ian 正文配图 `assets/success-rate-routing-illustrations/01-one-ranking-many-tracks.png` 为 1672×941，用于建立“总榜不是因果”的直觉。
- 2 个 Go 教学代码块覆盖固定查找、收缩估计与 Wilson 区间；正文明确说明教学代码“仅接受已知值、无先验”的简化，并映射生产实现中的修订、状态、阈值、待定结果、国家、网络、窗口与保守 PPM。

## 发布验证

- 本章 2 个 Go 代码块由正文提取为 1 个独立包，编译、竞态检测与静态检查已通过；附加测试覆盖修订、待定结果覆盖率与双年龄边界。
- 本地 Markdown 链接、图片路径、Mermaid、支付术语和固定实现链接已检查。
- 固定实现的单元测试、竞态检测、静态检查、OpenAPI、PostgreSQL 集成测试与 Compose 端到端测试，已由持续集成任务 `33041147119` 在固定提交上验证成功。

## 能力边界

当前快照、样本和结果都是合成夹具。表现包可校验 `authorize + manual` 统计模式，但当前 v3 只安装 `sale + automatic` 快照，手动请款资金链并未实现。代码没有真实 PSP 结果归一化与逻辑操作去重生产器，也没有流处理、特征存储、制品签名、审批与发布、动态健康度、实验或离线策略评估。它只能证明观察性读取合同，不能证明真实授权提升、因果收益、服务商基准表现或生产支付能力。
