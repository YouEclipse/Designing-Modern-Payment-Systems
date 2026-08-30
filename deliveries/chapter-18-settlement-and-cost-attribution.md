# 第 18 章《结算与成本归因》交付记录

## 章节契约

- 核心命题：预期成本与来源可验证的实际成本必须分层；结算只能追加经济事实和重建归因投影，不能把支付成功、银行到账与净额压成一个状态。
- 演进线：在线估费 → 文件与行身份及完整性 → 强关联匹配 → 守恒分配 → 校准提案。
- 本章代码基线：路由服务 v3 已生成版本化的条件预期成本证据；结算服务与实际事实尚不存在。
- 后续范围：制品发布和模型标签消费分别留给第 20、21 章。

## 能力状态

- `implemented`：固定实现中的版本化预期成本、精确连接、FX 与生效时间证据，以及支付优化快照。
- `teaching_model`：正文中的结算事实、强匹配与守恒分配 Go 模型。
- `target_contract`：服务商报告适配器、实际经济事实、异常队列、归因投影、校准发布器，以及账本与打款集成。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 访问边界：当前私有仓库证据需授权；不改变仓库可见性。
- 本章是设计型结算交付，没有虚构 `services/settlement` 运行时。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 条件预期成本分开支付尝试、成功与失败场景 | `services/router/internal/cost/estimate.go`；`types.go` | `TestFeeScenariosReproducePriceCrossingWithoutInventingProbability`；`TestCostUnknownAndOverflowNeverExposePartialZeroTotals` |
| `implemented` | 成本证据与精确连接修订版、FX、版本和有效期绑定 | `services/router/internal/cost/artifact.go`；`decision/optimization.go` | `TestOptimizationManifestAndClassificationBindExactRevision`；`TestFXHasExplicitMajorUnitDirectionAndExpiry` |
| `implemented` | 支付服务保存选择时的优化快照，历史估算不重算 | `services/payment/internal/postgres/routing_evidence.go`；迁移 `006` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |
| `teaching_model` | 结算行幂等、强匹配和分配守恒仍是教学合同 | 本章 `settlementfact`、`settlementmatch`、`costallocation` 包 | `scripts/go-examples/tests/chapter18` 编译、竞态与静态检查；重复、碰撞与守恒边界测试 |

## 图表与书稿验证

- Mermaid 区分授权、请款、清算、结算与打款，并展示仅追加的归因数据流。
- 本章 Go 示例以第 18 章专属测试做编译、竞态与静态检查。
- 无新增 Ian 插图；经济阶段与事实边界由精确图表承担。

## 发布验证

- 书稿发布前必须运行 `go run scripts/verify-go-examples.go part-4-operating-at-scale/settlement.md`、`scripts/verify-mermaid.sh part-4-operating-at-scale/settlement.md` 与 `git diff --check`；这些门禁只证明结算、匹配与分配教学模型。
- 当前运行时主张只能引用交付清单中列出的成本证据代码测试，并在 `bc4afde24da708d22f67da72a90a1557b1477173` 与 `https://github.com/YouEclipse/payments/actions/runs/33041147119` 被同一绿色 CI 固定后发布；不得据此宣称结算、账本或打款运行时已存在。

## 未实现与能力边界

- 没有 `services/settlement`、报告适配器、经济事实模式、异常队列、实际归因投影或校准发布器。
- 当前预期成本来自本地与测试制品，不是服务商账单、结算完成或生产费率证明。
- 支付服务仍只有自动请款 `SALE`；手动请款、撤销、退款、账本与打款均未实现。
