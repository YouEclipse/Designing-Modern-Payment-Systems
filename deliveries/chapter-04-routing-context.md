# 第 4 章《构建 Routing Context》交付记录

## 章节契约

- 核心命题：Router 的输入必须是一份由字段所有者构建、可验证且不含支付凭证的不可变事实快照，而不是上游领域对象的集合。
- 演进线：七字段函数参数 → 对象透传失控 → typed/provenanced Context → reader-first API → Attempt 事务内固定证据。
- 本章增量：`routing_context.v1`、Payment Builder、Router 严格 reader、共享 wire golden、在线 freshness 与 Context/Decision JSONB 快照。
- 后续范围：费用、成功率、Risk Action、动态 Health、Decision Store 与配置平台分别留给第 7～9、13、14、20 章。

## Status

- `implemented`：固定实现中的 `routing_context.v1` Builder/reader、严格 wire、freshness、摘要与 Attempt 事务内证据固定。
- `teaching_model`：正文 5 个 Go 块中的参数/对象反例、typed Context 与 Builder 模型。
- `target_contract`：真实 Merchant/Payment Method/Risk 来源、完整历史 Artifact、动态 Health、配置控制面与真实在线数据源；Cost/Performance 的当前状态以第 7～9 章 manifest 为准。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- CI：[33041147119 — success](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- API：Payment 显式选择 v2 writer；Router `/internal/v2/routing-decisions` 严格读取 `routing_context.v1`，失败时不自动回退 v1。
- 持久化：迁移 `004` 允许历史 Attempt 保持双 NULL，新执行则把 Context 与 Decision 在 Provider dispatch 前一起固定。

## 主张—实现—测试

工程文件相对于固定实现。

| Status | 主张 | 工程文件 | 验证 |
| --- | --- | --- | --- |
| `implemented` | Context 不携带 Token/PAN/CVV | Payment/Router `internal/routingcontext`；Router strict HTTP reader | `TestSnapshotIsDeterministicAndContainsNoCredentials`；`TestV2StrictHTTPBoundaryRejectsWithoutEchoingUntrustedInput` |
| `implemented` | 两端独立类型仍解释同一 wire | `contracts/internal/fixtures/routing-context-v1.json` | 两端 golden tests 与完整 SHA-256 digest |
| `implemented` | known/unknown/stale 不互相冒充 | 两端 Context validation 与 Payment Builder | `TestBuilderRejectsMissingStaleAndInvalidInputs`；`TestContextObservedStatesNeverCoerceToKnown` |
| `implemented` | Context 构建后不受输入别名修改 | Payment `routingcontext.Builder` | `TestBuilderDoesNotAliasMutableMetadataOrProfileInput` |
| `implemented` | 在线执行拒绝过期输入和配置 | Router v2 HTTP admission | `TestV2OnlineFreshnessAndMissingConfigNeverFallBackToV1`；跨到期边界复检测试 |
| `implemented` | 决策证据与 Intent、Attempt、连接绑定 | Payment `postgres/routing_evidence.go` 与迁移 `004` | unit 与 PostgreSQL integration immutable tests |
| `implemented` | `UNKNOWN` 恢复不重建 Context/Decision | Payment Operation 恢复路径；Compose flow | E2E response-loss 场景只有一个 Attempt 和一个 Provider Operation |

## 书稿验证

- 本章 5 个 Go 块由正文提取为 4 个独立 package，compile/race/vet 通过。
- 4 个 Mermaid 已实际渲染，覆盖对象膨胀、字段权威、Builder 边界与 reader-first 升级。
- 本章没有加入装饰性正文配图；结构关系由可精确解析的 Mermaid 表达。
- 本地链接、术语、固定实现链接与资源路径通过检查；固定实现的 unit/race/vet/OpenAPI/PostgreSQL integration/Compose E2E 门禁全绿。

## 能力边界

Local Profile 与 metadata 是有版本、会过期的 Simulator fixture，不是 Merchant Service、Vault 或生产 Risk 接入。当前 Context 只覆盖 tokenized card `SALE`/`AUTHORIZE` 的最小事实；摘要不是匿名化，保存 JSONB 也不等于拥有完整历史 Artifact 库。没有真实 PSP sandbox 或 live payment 能力。
