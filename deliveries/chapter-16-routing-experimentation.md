# 第 16 章《Routing Experimentation》交付记录

## 章节契约

- 核心命题：Assignment、Applied Decision、Provider Exposure 与 Outcome 是不同事实；实验只能在共同硬门禁内改变未来策略，不能生成第二条资金路径。
- 演进线：按请求随机 → Intent 稳定分桶 → assignment/exposure 分列 → shadow、SRM、成熟度与 kill switch。
- 本章代码基线：Router v3 已有受约束 Cost/Performance/Scoring 与 Payment 执行快照；experiment/assignment/shadow runtime 尚不存在。
- 后续范围：模型 OPE 与制品发布分别留给第 20、21 章。

## Status

- `implemented`：固定实现中的 Eligibility/Risk 硬门禁、champion 执行快照与 UNKNOWN 固定身份恢复基线。
- `teaching_model`：正文中的稳定 assignment、SRM 和 shadow 零副作用 Go 模型。
- `target_contract`：Intent assignment binding、experiment/shadow worker、kill switch、成熟标签、live ramp 与因果上线判定。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前 private 仓库证据仅对授权身份可见，不改变可见性。
- 本章是设计型交付，没有虚构 production experiment 增量。

## 主张—实现—验证映射

| Status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 当前 Cost/Performance/Scoring 只在 Eligibility/Rule/Risk 硬门禁之后运行，且不能复活已排除候选 | `services/router/internal/decision/optimization.go` | `TestEveryNonAllowRiskActionStopsBeforeFeatureEvaluation`；`TestOptimizationCannotReviveAnEligibilityOrRuleExclusion` |
| `target_contract` | 未来 experiment/assignment/shadow 也只能接在共同硬门禁之后 | 未来 experiment runtime 与 Router authority 合同 | 当前无 experiment runtime test；正文 authority 合同 |
| `implemented` | 当前 champion 决策已有完整可执行快照，但没有 assignment | `services/payment/internal/postgres/routing_evidence.go`；迁移 `006_risk_routing_evidence.sql` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |
| `teaching_model` | 稳定分桶、SRM 与 shadow 零副作用是教学合同，不是 runtime | 本章 Go 包 `experiment`、`srm`、`shadow` | `scripts/go-examples/tests/chapter16` compile/race/vet；稳定 1000 Intent golden、shadow no-operation tests |
| `implemented` | 当前 UNKNOWN 恢复始终沿固定 connection/operation，不重新分组 | Payment recovery 与 pinned identity | `TestUnknownBlocksNewConfirmationAndConverges`；`TestResponseLossRecoveryQueriesThenResubmitsTheSameDurableOperation` |
| `target_contract` | 未来 kill/promotion 不得改变已执行或 UNKNOWN Intent 的 assignment、Plan 与资金路径 | 未来 assignment binding / experiment runtime | 当前无 experiment runtime test；正文 authority 与 failure-mode 合同 |

## 图表与书稿验证

- Mermaid 覆盖实验阶段门禁、assignment 原子固定与 shadow 独立支路。
- 本章 Go 示例由正文提取，并以 chapter16 专属 tests 做 compile/race/vet。
- 无新增 Ian 插图；统计与执行权限关系由表格、公式和时序图表达。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-4-operating-at-scale/experimentation.md`、`scripts/verify-mermaid.sh part-4-operating-at-scale/experimentation.md` 与 `git diff --check`；这些门禁只证明 assignment/SRM/shadow 教学模型。
- 当前 runtime 主张只能引用 manifest 中列出的代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；没有 experiment/assignment/shadow runtime 完成声明。

## 未实现与能力边界

- 无 `internal/experiment`、Intent assignment binding、shadow worker、live ramp、kill-switch state 或成熟度 scorecard。
- Router v3 的 observational Performance 不是随机实验，也不提供 propensity 或因果提升证明。
- 未运行真实 A/B，更没有用 Simulator 结果宣称 live 收益。
