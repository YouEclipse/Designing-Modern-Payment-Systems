# 第 4 章《构建路由上下文》交付记录

## 章节契约

- 核心命题：路由服务的输入必须是一份由字段所有者构建、可验证且不含支付凭证的不可变事实快照，而不是上游领域对象的集合。
- 演进线：七字段函数参数 → 对象透传失控 → 有类型、有来源的上下文 → 先部署读取方的接口升级 → 支付尝试事务内固定证据。
- 本章增量：`routing_context.v1`、支付侧构建器、路由侧严格读取器、共享线格式黄金样例、在线新鲜度，以及上下文与决策的 JSONB 快照。
- 后续范围：费用、成功率、风险动作、动态健康度、决策存储与配置平台分别留给第 7～9、13、14、20 章。

## 能力状态

- `implemented`：固定实现中的 `routing_context.v1` 构建器与读取器、严格线格式、新鲜度、摘要与支付尝试事务内证据固定。
- `teaching_model`：正文 5 个 Go 代码块中的参数与对象反例、有类型的上下文与构建器模型。
- `target_contract`：真实商户、支付方式和风险来源，完整历史制品，动态健康度、配置控制面与真实在线数据源；成本与表现的当前状态以第 7～9 章交付记录为准。

## 配套实现

- 固定实现：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)。
- 持续集成：[33041147119 — 成功](https://github.com/YouEclipse/payments/actions/runs/33041147119)。
- 接口：支付服务显式选择 v2 写入方；路由服务 `/internal/v2/routing-decisions` 严格读取 `routing_context.v1`，失败时不自动回退 v1。
- 持久化：迁移 `004` 允许历史支付尝试保持双 `NULL`，新执行则在派发给支付服务商前一起固定上下文与决策。

## 主张—实现—测试

工程文件相对于固定实现。

| 状态 | 主张 | 工程文件 | 验证 |
| --- | --- | --- | --- |
| `implemented` | 路由上下文不携带令牌、PAN 或 CVV | 支付侧与路由侧的 `internal/routingcontext`；路由服务的严格 HTTP 读取器 | `TestSnapshotIsDeterministicAndContainsNoCredentials`；`TestV2StrictHTTPBoundaryRejectsWithoutEchoingUntrustedInput` |
| `implemented` | 两端独立类型仍解释同一线格式 | `contracts/internal/fixtures/routing-context-v1.json` | 两端黄金样例测试与完整 SHA-256 摘要 |
| `implemented` | 已知、未知与过期状态不互相冒充 | 两端上下文校验与支付侧构建器 | `TestBuilderRejectsMissingStaleAndInvalidInputs`；`TestContextObservedStatesNeverCoerceToKnown` |
| `implemented` | 上下文构建后不受输入别名修改 | 支付侧 `routingcontext.Builder` | `TestBuilderDoesNotAliasMutableMetadataOrProfileInput` |
| `implemented` | 在线执行拒绝过期输入和配置 | 路由服务 v2 HTTP 准入 | `TestV2OnlineFreshnessAndMissingConfigNeverFallBackToV1`；跨到期边界复检测试 |
| `implemented` | 决策证据与支付意图、支付尝试、连接绑定 | 支付侧 `postgres/routing_evidence.go` 与迁移 `004` | 单元测试与 PostgreSQL 不可变性集成测试 |
| `implemented` | `UNKNOWN` 恢复不重建上下文或决策 | 支付资金操作恢复路径；Compose 流程 | 响应丢失端到端场景只有一个支付尝试和一项服务商资金操作 |

## 书稿验证

- 本章 5 个 Go 代码块由正文提取为 4 个独立包，编译、竞态检测和静态检查通过。
- 4 个 Mermaid 已实际渲染，覆盖对象膨胀、字段权威、构建器边界与先部署读取方的升级顺序。
- 本章没有加入装饰性正文配图；结构关系由可精确解析的 Mermaid 表达。
- 本地链接、术语、固定实现链接与资源路径通过检查；固定实现的单元测试、竞态检测、静态检查、OpenAPI、PostgreSQL 集成测试和 Compose 端到端测试门禁全部通过。

## 能力边界

本地画像与元数据是有版本、会过期的模拟器夹具，不是商户服务、凭证保险库或生产风险接入。当前路由上下文只覆盖令牌化银行卡 `SALE` 与 `AUTHORIZE` 的最小事实；摘要不是匿名化，保存 JSONB 也不等于拥有完整历史制品库。系统没有真实 PSP 沙箱或生产支付能力。
