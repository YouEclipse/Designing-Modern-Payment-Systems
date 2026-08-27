# 第 5 章《Rule-Based Routing》交付记录

## 章节契约

- 核心命题：规则只能在已允许集合内单调收缩或重排；Rule Set 必须先编译成不可变、确定且可解释的 Artifact。
- 演进线：Go 分支顺序 → 声明式条件的歧义 → 受限 AST 与三值真值 → 冲突拒绝、canonical digest 与稳定分桶。
- 本章增量：`routing_rules.v1` compiler/evaluator，`force`、`exclude`、`prefer`、`split` 以及 Decision 中的 Rule Hits、version/digest。
- 后续范围：历史版本库、审批、热发布、shadow 调度与跨 Region ACK 留给第 16、20 章。

## Status

- `implemented`：固定实现中的 `routing_rules.v1` 编译/求值、三值条件、稳定分桶、冲突拒绝与 Decision 证据。
- `teaching_model`：正文 3 个 Go 块中的分支反例与最小规则编译/求值模型。
- `target_contract`：历史 Artifact Store、审批、热发布、shadow 调度、跨 Region activation 与真实 Provider 配置。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[33041147119 — success](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 纵向切片：Router v2 从本地不可变 Rule Set 计算计划和解释；无候选返回带证据的 `422`，Payment 不创建资金操作。

## 主张—实现—测试

工程文件相对于固定实现。

| Status | 主张 | 工程文件 | 验证 |
| --- | --- | --- | --- |
| `implemented` | Unknown 不能因为 `!=` 或缺值变成 True | `services/router/internal/rules` typed condition evaluator | `TestRuleTruthUsesTypedFieldsAndKleeneConjunction` |
| `implemented` | Action 不能复活被硬约束排除的连接 | rules evaluator + decision pipeline | `TestForceExcludeAndSplitCannotResurrectIneligibleCandidate`；pipeline 对照测试 |
| `implemented` | 同优先级可证明冲突在安装前失败 | rules compiler overlap validation | 冲突、范围边界与 disjoint 对照测试 |
| `implemented` | Artifact 内容、集合和求值顺序确定 | canonical compiler/digest | `TestRuleCanonicalArtifactsImmutableAndOrderIndependent` |
| `implemented` | Split 只在允许成员中稳定分配 | rules split evaluator | `TestSplitStableAssignmentOnlyWithinEligibleWeightedSet` |
| `implemented` | caller 不能提交 Rule、版本或 Token | Router strict HTTP boundary | v1/v2 unknown-field、token rejection tests |
| `implemented` | 原 Operation 恢复不重新执行 Rule | Payment 固定 Operation identity | Compose `UNKNOWN` 收敛检查一个资金路径 |

## 书稿验证

- 本章 3 个 Go 块由正文提取为 2 个独立 package，compile/race/vet 通过；附加测试覆盖 Kleene 真值与 preference 硬边界。
- 4 个 Mermaid 已实际渲染，覆盖分支膨胀、Rule pipeline、shadow 与紧急禁用。
- 本章没有用正文插图代替可审计规则图。
- 本地链接、固定实现链接、支付术语与跨章边界通过检查；固定实现的完整 CI 门禁全绿。

## 能力边界

当前实现没有远程脚本、历史 Artifact Store、在线审批、实验分流服务或配置热更新。Local Rule Set 是 Simulator fixture；version/digest 只能识别已加载内容，不能单独证明发布审批或防止控制面复用版本。没有真实 PSP sandbox 或 live payment 能力。
