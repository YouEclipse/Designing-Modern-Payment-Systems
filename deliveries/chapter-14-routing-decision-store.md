# 第 14 章《路由决策存储》交付记录

## 章节契约

- 核心命题：历史选择必须由决策时封存的不可变事实回答；支付服务的本地执行快照是执行权威，远端存储只能异步追加和查询。
- 演进线：`selected_provider` 日志 → 完整 v3 执行快照 → 规范源字节与完整哈希 → 发件箱 → 只追加投影与隔离区。
- 当前代码增量：完整上下文、决策、风险、优化、候选顺序、`executable_until` 与支付尝试、资金操作原子固定且不可改写。
- 后续范围：规范字节与完整 SHA 协议、决策专用发件箱、决策存储与投影器，以及 `RoutePlanDecision` 和支付服务编写的 `FallbackExecutionBinding` 后继事务尚未实现。

## 配套实现

- 代码仓库：`YouEclipse/payments`
- 实现提交：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- 持续集成任务：[33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 纵向切片：支付服务在首次资金输入输出前，以事务固定完整路由服务 v3 JSON 证据与截止时间；JSONB 快照是本地执行真相，但不是最终规范档案。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 路由服务与支付服务对同一 v3 上下文和决策黄金样例无损解析 | 契约夹具；路由侧 `internal/httpapi/v2.go`；支付侧 `clients/router_v3_test.go` | `TestV3HTTPSuccessMatchesSharedContractGolden`、`TestRouterV3ConsumesSharedRouterSuccessGoldenWithoutLoss` |
| `implemented` | 上下文主体与摘要、风险 `ALLOW`、优化顺序、选中修订与截止时间在事务内复验 | `services/payment/internal/postgres/routing_evidence.go` | `TestRoutingEvidenceBindsLockedIntentAndSelectedExecutionIdentity`、`TestRouterV3RejectsTamperedOrCredentialLikeEvidence` |
| `implemented` | 完整上下文与决策快照、支付尝试、资金操作与截止时间原子固定，之后不可改写 | 迁移 `006_risk_routing_evidence.sql`；`services/payment/internal/postgres/store.go` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |
| `implemented` | 首次资金输入输出超过 `executable_until` 边界后拒绝执行；已派发操作的恢复不会被错误重路由 | 支付恢复与存储 | `TestRoutingEvidenceExpiresBeforeFirstProviderDispatch`、`TestExpiredRoutingAuthorityDoesNotBlockAlreadyDispatchedSameKeyRecovery` |
| `implemented` | 支付服务从 JSONB 重读后，按有类型的优化线格式语义重建 `rtd_` 材料：对象键重排仍通过，字段值或有序数组变化会失败，含糊 JSON 会被拒绝；该编号仍只是截断关联编号 | `services/payment/internal/clients/router_optimization.go`；路由决策编码器 | `TestVersionedDecisionContentIDUsesOptimizationSemanticsAcrossObjectReordering`、`TestVersionedDecisionContentIDRejectsAmbiguousOptimizationJSON`；共享黄金样例与篡改测试 |
| `teaching_model` | 规范信封用 `Outcome` 判别联合区分已选结果与六类负向结果；负向结果不伪造支付尝试、计划或选择，可保存 `not_exposed` 分组归属 | 本章 Go 代码块 | 第 14 章 `envelope_test.go` 的结果分支、仅分组归属与矛盾测试 |
| `teaching_model` | 固定与投影的教学合同可编译、可反例验证 | 本章 Go 代码块 | 第 14 章 `pinning_test.go`、`projection_test.go` |
| `implemented` | 两个已请款结果保留业务成功并独立进入完整性复核，不覆盖路由计划或支付侧后备绑定 | 支付领域模型 | `TestConflictingTerminalOutcomeRequiresReview`、`TestReviewRequiredIsNotIgnoredAfterTerminalSuccess` |

## 图表与配图

- 4 个 Mermaid：历史事实与反事实的差异、本地快照与发件箱、摄取冲突、后备路由决策谱系。
- Ian 正文配图：`assets/routing-decision-store-illustrations/01-the-decision-black-box.png`，1672×941；表达“箱内历史不随当前配置变化”。
- 4 个 Go 代码块由正文提取，第 14 章测试覆盖已选结果与负向结果的判别联合、仅分组归属、规范信封、事务固定与投影摄取。

## 发布验证

- 书稿侧执行章节 Go 编译、竞态检测、静态检查、Mermaid 实际渲染、本地链接、图片路径与尺寸检查，以及 `git diff --check`。
- v3 本地执行快照的单元测试、集成测试和 Compose 证据以固定持续集成任务为准；持续集成成功不得被解释为远端决策存储已经实现。

## 未实现与后续章节

- JSONB 不保存原始规范字节；当前有类型的语义重建只用于路由服务与支付服务现有的内容编号绑定，`rtd_` 仍只有截断编号，没有原始规范档案、持久化完整 SHA、规范化版本、源记录或跨语言规范化协议。
- 无 `routing_decision.recorded` 发件箱、只追加源数据、投影与隔离存储，也没有投影器或重建流程。
- 无专用后备后继事务；未来由支付服务编写的 `FallbackExecutionBinding` 只能绑定路由服务编写的 `RoutePlanDecision` 固定下一项，以及新鲜风险证据与护栏证据。路由服务所有的复检只能否决，不能产生新候选集合或排序、重排或续期；证据过期或能力不可用时必须停止。
