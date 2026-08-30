# 第 5 章《基于规则的路由》交付记录

## 章节契约

- 核心命题：规则只能在已允许集合内单调收缩或重排；规则集必须先编译成不可变、确定且可解释的制品。
- 演进线：Go 分支顺序 → 声明式条件的歧义 → 受限抽象语法树与三值真值 → 冲突拒绝、规范摘要与稳定分桶。
- 本章增量：`routing_rules.v1` 编译器与求值器，`force`、`exclude`、`prefer`、`split`，以及决策中的规则命中、版本和摘要。
- 后续范围：历史版本库、审批、热发布、影子调度与跨区域确认留给第 16、20 章。

## 能力状态

- `implemented`：固定实现中的 `routing_rules.v1` 编译与求值、三值条件、稳定分桶、冲突拒绝与决策证据。
- `teaching_model`：正文 3 个 Go 代码块中的分支反例与最小规则编译、求值模型。
- `target_contract`：历史制品存储、审批、热发布、影子调度、跨区域启用与真实支付服务商配置。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- 持续集成：[33041147119 — 成功](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 纵向切片：路由服务 v2 从本地不可变规则集计算计划和解释；无候选时返回带证据的 `422`，支付服务不创建资金操作。

## 主张—实现—测试

工程文件相对于固定实现。

| 状态 | 主张 | 工程文件 | 验证 |
| --- | --- | --- | --- |
| `implemented` | 未知值不能因为 `!=` 或字段缺失变成真 | `services/router/internal/rules` 有类型的条件求值器 | `TestRuleTruthUsesTypedFieldsAndKleeneConjunction` |
| `implemented` | 动作不能复活被硬约束排除的连接 | 规则求值器与决策流水线 | `TestForceExcludeAndSplitCannotResurrectIneligibleCandidate`；流水线对照测试 |
| `implemented` | 同优先级可证明冲突在安装前失败 | 规则编译器的重叠校验 | 冲突、范围边界与不相交对照测试 |
| `implemented` | 制品内容、集合和求值顺序确定 | 规范编译器与摘要 | `TestRuleCanonicalArtifactsImmutableAndOrderIndependent` |
| `implemented` | 分流只在允许成员中稳定分配 | 规则分流求值器 | `TestSplitStableAssignmentOnlyWithinEligibleWeightedSet` |
| `implemented` | 调用方不能提交规则、版本或令牌 | 路由服务的严格 HTTP 边界 | v1 与 v2 的未知字段、令牌拒绝测试 |
| `implemented` | 恢复原资金操作时不重新执行规则 | 支付服务固定资金操作身份 | Compose `UNKNOWN` 收敛检查只有一条资金路径 |

## 书稿验证

- 本章 3 个 Go 代码块由正文提取为 2 个独立包，编译、竞态检测和静态检查通过；附加测试覆盖 Kleene 三值逻辑与偏好硬边界。
- 4 个 Mermaid 已实际渲染，覆盖分支膨胀、规则流水线、影子求值与紧急禁用。
- 本章没有用正文插图代替可审计规则图。
- 本地链接、固定实现链接、支付术语与跨章边界通过检查；固定实现的完整持续集成门禁全部通过。

## 能力边界

当前实现没有远程脚本、历史制品存储、在线审批、实验分流服务或配置热更新。本地规则集是模拟器夹具；版本与摘要只能识别已加载内容，不能单独证明发布审批或防止控制面复用版本。系统没有真实 PSP 沙箱或生产支付能力。
