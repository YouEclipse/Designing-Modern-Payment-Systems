# 第 18 章《Settlement 与 Cost Attribution》交付记录

## 章节契约

- 核心命题：Expected Cost 与来源可验证的 Actual Cost 必须分层；Settlement 只能追加经济事实和重建归因投影，不能把支付成功、银行到账与净额压成一个状态。
- 演进线：在线估费 → 文件/行身份与完整性 → 强关联匹配 → 守恒分配 → calibration proposal。
- 本章代码基线：Router v3 已生成版本化条件 Expected Cost evidence；Settlement service 与 Actual facts 尚不存在。
- 后续范围：制品发布和模型标签消费分别留给第 20、21 章。

## Status

- `implemented`：固定实现中的版本化 Expected Cost、精确 connection/FX/effective-time 证据与 Payment Optimization snapshot。
- `teaching_model`：正文中的 settlement fact、强匹配与守恒分配 Go 模型。
- `target_contract`：Provider report adapter、Actual economic facts、exception queue、attribution projection、calibration publisher、Ledger 与 Payout 集成。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前 private 仓库证据需授权；不改变仓库可见性。
- 本章是设计型 Settlement 交付，没有虚构 `services/settlement` runtime。

## 主张—实现—验证映射

| Status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 条件 Expected Cost 分开 attempt/success/failure scenario | `services/router/internal/cost/estimate.go`；`types.go` | `TestFeeScenariosReproducePriceCrossingWithoutInventingProbability`；`TestCostUnknownAndOverflowNeverExposePartialZeroTotals` |
| `implemented` | Cost evidence 与 exact connection revision、FX、版本和有效期绑定 | `services/router/internal/cost/artifact.go`；`decision/optimization.go` | `TestOptimizationManifestAndClassificationBindExactRevision`；`TestFXHasExplicitMajorUnitDirectionAndExpiry` |
| `implemented` | Payment 保存选择时的 Optimization snapshot，历史 estimate 不重算 | `services/payment/internal/postgres/routing_evidence.go`；迁移 `006` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |
| `teaching_model` | Settlement 行幂等、强匹配和分配守恒仍是教学合同 | 本章 `settlementfact`、`settlementmatch`、`costallocation` 包 | `scripts/go-examples/tests/chapter18` compile/race/vet；重复/碰撞/守恒边界测试 |

## 图表与书稿验证

- Mermaid 区分 Authorization/Capture/Clearing/Settlement/Payout，并展示 append-only attribution data flow。
- 本章 Go 示例以 chapter18 专属 tests 做 compile/race/vet。
- 无新增 Ian 插图；经济阶段与事实边界由精确图表承担。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-4-operating-at-scale/settlement.md`、`scripts/verify-mermaid.sh part-4-operating-at-scale/settlement.md` 与 `git diff --check`；这些门禁只证明 settlement/matching/allocation 教学模型。
- 当前 runtime 主张只能引用 manifest 中列出的 Cost evidence 代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；不得据此宣称 Settlement、Ledger 或 Payout runtime。

## 未实现与能力边界

- 没有 `services/settlement`、report adapter、经济 facts schema、exception queue、Actual attribution projection 或 calibration publisher。
- 当前 Expected Cost 来自 local/test Artifact，不是 Provider invoice、结算完成或 live 费率证明。
- Payment 仍只有 automatic-capture SALE；manual Capture/Void/Refund、Ledger/Payout 均未实现。
