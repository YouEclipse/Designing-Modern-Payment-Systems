# 第 13 章《Circuit Breaker 与 Provider Health》交付记录

## 章节契约

- 核心命题：Health 是按 exact Candidate/revision/operation/fault-domain 聚合的未来流量证据，不能裁决当前 UNKNOWN 或授权 Fallback。
- 演进线：全局 `healthy bool` → typed Technical Observation → minimum-sample window → Unavailable/Recovering → shared probe 与 per-key Bulkhead。
- 当前代码基线：Router v3 constrained pipeline 与 Gateway append-only Provider Operation evidence 已实现。
- 明确设计项：Technical Health ingest/store/window/snapshot/probe/bulkhead 尚未实现。

## 配套实现

- Repository：`YouEclipse/payments`
- Implementation commit：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- CI run：[https://github.com/YouEclipse/payments/actions/runs/33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 本章纵向切片：本轮交付可编译、可反例验证的 Health/Bulkhead 教学模型；真实微服务中仅提供未来 Health evidence 可接入的 v3 决策骨架。

## 主张—实现—验证映射

| status | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | Router 现状是 v3 Eligibility→Rule/Risk→Cost/Performance→Scoring，而非静态排序 | `services/router/internal/decision/optimization.go`；`cost/`、`performance/`、`scoring/` | `TestV3HTTPSuccessMatchesSharedContractGolden`、Optimization/scoring package tests |
| `implemented` | Payment 固定 v3 evidence，不会用当前配置重算已 dispatch Operation | `services/payment/internal/postgres/routing_evidence.go` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |
| `teaching_model` | 教学 Health state machine 验证 previous State enum，Unavailable 必须有合法非零 OpenUntil | 本章 `package health` Go block | `TestInvalidPreviousStateOrUnavailableWindowFailsClosed` |
| `teaching_model` | Issuer/business response 不计 technical failure；minimum sample、迟滞与 failed probe 可证伪 | 本章 `package health` Go blocks | chapter13 `health_edges_test.go`、`TestIssuerDeclineDoesNotOpenAndFailedProbeReopens` |
| `teaching_model` | Bulkhead permit 可取消、释放幂等并限制最大并发；Query 可有独立容量 | 本章 `package bulkhead` Go block | chapter13 `bulkhead_test.go` |
| `implemented` | 当前 UNKNOWN 始终固定原 connection；Observation ledger 不等于 Health aggregate | Payment Recovery 与 Gateway observations | `TestSameProviderConnectionsStayPinnedThroughUnknownRecovery`、`TestPostgresProviderObservationLedgerFencesStaleGenerationAndIsImmutable` |

## 图表与配图

- 4 个 Mermaid：单次 timeout 双事实、Health state、shared probe、异步 observation pipeline。
- 本章未添加 Ian 插图；状态机、时序和故障域表承担主要解释。
- 5 个 Go 块由正文提取，chapter13 tests 覆盖 naive breaker、分类/窗口/状态与 Bulkhead 并发。

## 发布验证

- 书稿侧执行章节 Go compile/race/vet、Mermaid 实际渲染、本地链接与 `git diff --check`。
- Router v3/Gateway evidence 的代码证据以固定 CI run 为准；Health 教学模型通过不等于生产 Health subsystem 已存在。

## 未实现与后续章节

- 无 `services/router/internal/health`、Health PostgreSQL migration/ingest、Snapshot publisher、shared probe lease 或 Gateway per-connection/operation Bulkhead。
- 无真实 PSP health 数据、Dashboard、自动 incident override 或局部故障 Compose E2E。
