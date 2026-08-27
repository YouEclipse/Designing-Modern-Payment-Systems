# 第 21 章《Intelligent Routing》交付记录

## 章节契约

- 核心命题：模型只是合法 Candidate Set 内可替换的软估计器；没有完整 support、可信 propensity 和成熟标签，就不能宣称因果提升。
- 演进线：预测成功率 → dataset/support 审计 → OPE fail-closed → 有界 Artifact、whole-batch fallback 与安全探索。
- 本章代码基线：Router v3 已有 observational Performance + Cost Scoring；Model/OPE/Feature pipeline 尚未实现。
- 后续范围：全平台容量和上线验收留给第 22 章。

## Status

- `implemented`：固定实现中的 observational Performance、Cost/Scoring、Eligibility/Risk authority 与 whole-batch unavailable fallback。
- `teaching_model`：正文中的 OPE support 检查、model artifact loader 与 modelbatch whole-batch fallback Go 模型。
- `target_contract`：Feature/label pipeline、训练与 registry、shadow integration、真实 propensity/OPE、模型发布、live experiment 与运营门禁。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前 private 仓库证据需授权；不改变可见性。
- 本章是设计型交付，没有把 observational scoring 冒充 ML runtime 或因果实验。

## 主张—实现—验证映射

| Status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | historical Performance 明确是 policy-conditional、非因果 | `services/router/internal/performance` | `TestShrinkageDoesNotTurnOneSuccessIntoCertainSuperiority`；`TestPerformanceCohortAndRevisionCannotSilentlyMix` |
| `implemented` | 当前 observational Performance/Cost/Scoring 不能复活 Eligibility/Rule/Risk 排除 | `services/router/internal/decision/optimization.go`；`services/router/internal/scoring/score.go` | `TestOptimizationCannotReviveAnEligibilityOrRuleExclusion`；`TestEligibilityIsAuthoritativeEvenWhenExcludedCandidateHasBestScore` |
| `target_contract` | 未来模型估计器同样不得复活任何硬门禁排除 | 未来 model integration 与 Router authority 合同 | 当前无 model runtime test；正文 authority 合同 |
| `implemented` | 当前 Cost/Performance/cross-border evidence 任一 unavailable 时，Scoring 只能按 profile 对整批 static fallback 或 reject，不保留部分数值 | `services/router/internal/decision/optimization.go`；`services/router/internal/scoring/score.go` | `TestUnavailableFeatureIsAWholeBatchAction` |
| `teaching_model` | 模型失败或部分预测也必须整批回到批准 profile | 本章 `modelbatch` Go 包 | chapter21 whole-batch fallback、NaN、timeout tests |
| `teaching_model` | Support/OPE/Artifact loader 必须 fail closed | 本章 `ope`、`modelartifact` Go 包 | `scripts/go-examples/tests/chapter21` compile/race/vet；zero-support/expiry tests |

## 图表与书稿验证

- Mermaid 覆盖训练—发布—执行边界、whole-batch fallback 与 model timeout。
- 本章 Go 示例以 chapter21 专属 tests 做 compile/race/vet。
- 无新增 Ian 插图；支持集合和概率边界通过公式、表格与流程图表达。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-5-global-routing/intelligent-routing.md`、`scripts/verify-mermaid.sh part-5-global-routing/intelligent-routing.md` 与 `git diff --check`；这些门禁只证明 OPE/model-artifact/whole-batch 教学模型。
- 当前 runtime 主张只能引用 manifest 中列出的 observational Performance/Scoring 代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；不得推导 live ML、因果收益或自治资金权限。

## 未实现与能力边界

- 无 `internal/model`、Feature pipeline、训练/registry、shadow integration、logging propensity、成熟标签或 `tools/routing-replay`。
- v3 Performance evidence 只反映历史策略下的观察，不是反事实成功率，也不授权探索。
- 没有 live model、A/B 因果收益或自治强化学习资金权限。
