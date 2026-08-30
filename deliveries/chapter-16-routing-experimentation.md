# 第 16 章《路由实验》交付记录

## 章节契约

- 核心命题：实验分配、实际决策、服务商暴露与结果是不同事实；实验只能在共同硬门禁内改变未来策略，不能生成第二条资金路径。
- 演进线：按请求随机 → 支付意图稳定分桶 → 分配与暴露分列 → 影子实验、SRM、成熟度与紧急停用开关。
- 本章代码基线：路由服务 v3 已有受约束的成本、性能与评分，以及支付执行快照；实验、分配与影子运行时尚不存在。
- 后续范围：模型 OPE 与制品发布分别留给第 20、21 章。

## 能力状态

- `implemented`：固定实现中的准入筛选与风险硬门禁、主方案执行快照与 `UNKNOWN` 固定身份恢复基线。
- `teaching_model`：正文中的稳定分配、SRM 和影子实验零副作用 Go 模型。
- `target_contract`：支付意图分配绑定、实验与影子工作进程、紧急停用开关、成熟标签、生产渐进放量与因果上线判定。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前私有仓库证据仅对授权身份可见，不改变可见性。
- 本章是设计型交付，没有虚构生产实验增量。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 当前成本、性能与评分只在准入筛选、规则和风险硬门禁之后运行，且不能复活已排除候选通道 | `services/router/internal/decision/optimization.go` | `TestEveryNonAllowRiskActionStopsBeforeFeatureEvaluation`；`TestOptimizationCannotReviveAnEligibilityOrRuleExclusion` |
| `target_contract` | 未来实验、分配和影子运行也只能接在共同硬门禁之后 | 未来实验运行时与路由服务权限合同 | 当前无实验运行时测试；正文权限合同 |
| `implemented` | 当前主方案决策已有完整可执行快照，但没有实验分配 | `services/payment/internal/postgres/routing_evidence.go`；迁移 `006_risk_routing_evidence.sql` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |
| `teaching_model` | 稳定分桶、SRM 与影子实验零副作用是教学合同，不是运行时 | 本章 Go 包 `experiment`、`srm`、`shadow` | `scripts/go-examples/tests/chapter16` 编译、竞态与静态检查；稳定 1000 个支付意图的黄金结果、影子实验无操作测试 |
| `implemented` | 当前 `UNKNOWN` 恢复始终沿固定连接与操作，不重新分组 | 支付恢复与固定身份 | `TestUnknownBlocksNewConfirmationAndConverges`；`TestResponseLossRecoveryQueriesThenResubmitsTheSameDurableOperation` |
| `target_contract` | 未来紧急停止与方案晋级不得改变已执行或 `UNKNOWN` 支付意图的分配、计划与资金路径 | 未来分配绑定与实验运行时 | 当前无实验运行时测试；正文权限与故障模式合同 |

## 图表与书稿验证

- Mermaid 覆盖实验阶段门禁、分配原子固定与影子实验独立支路。
- 本章 Go 示例由正文提取，并以第 16 章专属测试做编译、竞态与静态检查。
- 无新增 Ian 插图；统计与执行权限关系由表格、公式和时序图表达。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-4-operating-at-scale/experimentation.md`、`scripts/verify-mermaid.sh part-4-operating-at-scale/experimentation.md` 与 `git diff --check`；这些门禁只证明分配、SRM 与影子实验教学模型。
- 当前运行时主张只能引用交付清单中列出的代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；没有实验、分配或影子运行时完成声明。

## 未实现与能力边界

- 无 `internal/experiment`、支付意图分配绑定、影子工作进程、生产渐进放量、紧急停用状态或成熟度计分卡。
- 路由服务 v3 的观察性性能不是随机实验，也不提供倾向概率或因果提升证明。
- 未运行真实 A/B 实验，更没有用模拟器结果宣称生产收益。
