# 第 21 章《智能路由》交付记录

## 章节契约

- 核心命题：模型只是合法候选集内可替换的软估计器；没有完整支持范围、可信倾向概率和成熟标签，就不能宣称因果提升。
- 演进线：预测成功率 → 数据集与支持范围审计 → OPE 缺证据即拒绝 → 有界制品、整批后备与安全探索。
- 本章代码基线：路由服务 v3 已有观察性性能与成本评分；模型、OPE 和特征流水线尚未实现。
- 后续范围：全平台容量和上线验收留给第 22 章。

## 能力状态

- `implemented`：固定实现中的观察性性能、成本与评分、准入筛选和风险权威，以及不可用时的整批后备处理。
- `teaching_model`：正文中的 OPE 支持检查、模型制品加载器与整批后备 Go 模型。
- `target_contract`：特征与标签流水线、训练与注册表、影子集成、真实倾向概率与 OPE、模型发布、生产实验与运营门禁。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前私有仓库证据需授权；不改变可见性。
- 本章是设计型交付，没有把观察性评分冒充机器学习运行时或因果实验。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 历史性能明确受策略条件影响，不是因果结论 | `services/router/internal/performance` | `TestShrinkageDoesNotTurnOneSuccessIntoCertainSuperiority`；`TestPerformanceCohortAndRevisionCannotSilentlyMix` |
| `implemented` | 当前观察性性能、成本与评分不能复活被准入筛选、规则或风险排除的候选通道 | `services/router/internal/decision/optimization.go`；`services/router/internal/scoring/score.go` | `TestOptimizationCannotReviveAnEligibilityOrRuleExclusion`；`TestEligibilityIsAuthoritativeEvenWhenExcludedCandidateHasBestScore` |
| `target_contract` | 未来模型估计器同样不得复活任何硬门禁排除 | 未来模型集成与路由服务权限合同 | 当前无模型运行时测试；正文权限合同 |
| `implemented` | 当前成本、性能或跨境证据任一不可用时，评分只能按配置方案对整批使用静态后备或拒绝，不能保留部分数值 | `services/router/internal/decision/optimization.go`；`services/router/internal/scoring/score.go` | `TestUnavailableFeatureIsAWholeBatchAction` |
| `teaching_model` | 模型失败或部分预测也必须整批回到已批准配置方案 | 本章 `modelbatch` Go 包 | 第 21 章整批后备、NaN 与超时测试 |
| `teaching_model` | 支持范围、OPE 与制品加载器必须缺证据即拒绝 | 本章 `ope`、`modelartifact` Go 包 | `scripts/go-examples/tests/chapter21` 编译、竞态与静态检查；零支持与过期测试 |

## 图表与书稿验证

- Mermaid 覆盖训练—发布—执行边界、整批后备与模型超时。
- 本章 Go 示例以第 21 章专属测试做编译、竞态与静态检查。
- 无新增 Ian 插图；支持集合和概率边界通过公式、表格与流程图表达。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-5-global-routing/intelligent-routing.md`、`scripts/verify-mermaid.sh part-5-global-routing/intelligent-routing.md` 与 `git diff --check`；这些门禁只证明 OPE、模型制品与整批处理教学模型。
- 当前运行时主张只能引用交付清单中列出的观察性性能与评分代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；不得推导生产机器学习、因果收益或自治资金权限。

## 未实现与能力边界

- 无 `internal/model`、特征流水线、训练与注册表、影子集成、日志倾向概率、成熟标签或 `tools/routing-replay`。
- v3 性能证据只反映历史策略下的观察，不是反事实成功率，也不授权探索。
- 没有生产模型、A/B 因果收益或自治强化学习资金权限。
