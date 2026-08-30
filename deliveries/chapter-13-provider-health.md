# 第 13 章《熔断器与服务商健康度》交付记录

## 章节契约

- 核心命题：健康度是按精确候选通道、修订、资金操作与故障域聚合的未来流量证据，不能裁决当前 `UNKNOWN` 或授权后备路由。
- 演进线：全局 `healthy bool` → 有类型的技术观察 → 最小样本窗口 → 不可用或恢复中 → 共享探测与按键舱壁隔离。
- 当前代码基线：路由服务 v3 约束流水线与网关追加式服务商资金操作证据已经实现。
- 明确设计项：技术健康度摄取、存储、窗口、快照、探测与舱壁隔离尚未实现。

## 配套实现

- 代码仓库：`YouEclipse/payments`
- 实现提交：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- 持续集成任务：[33041147119](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 本章纵向切片：本轮交付可编译、可反例验证的健康度与舱壁隔离教学模型；真实微服务中仅提供未来健康度证据可接入的 v3 决策骨架。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 路由服务现状是“v3 准入筛选→规则与风险→成本与表现→评分”，而非静态排序 | `services/router/internal/decision/optimization.go`；`cost/`、`performance/`、`scoring/` | `TestV3HTTPSuccessMatchesSharedContractGolden`、优化与评分包测试 |
| `implemented` | 支付服务固定 v3 证据，不会用当前配置重算已派发资金操作 | `services/payment/internal/postgres/routing_evidence.go` | `TestRoutingEvidencePersistsBeforeDispatchAndCannotBeRewritten` |
| `teaching_model` | 教学健康度状态机验证前一状态枚举；不可用状态必须有合法且非零的开放截止时间 | 本章 `package health` Go 代码块 | `TestInvalidPreviousStateOrUnavailableWindowFailsClosed` |
| `teaching_model` | 发卡行与业务响应不计为技术失败；最小样本、迟滞与失败探测均可证伪 | 本章 `package health` Go 代码块 | 第 13 章 `health_edges_test.go`、`TestIssuerDeclineDoesNotOpenAndFailedProbeReopens` |
| `teaching_model` | 舱壁许可可取消、释放幂等并限制最大并发；查询可有独立容量 | 本章 `package bulkhead` Go 代码块 | 第 13 章 `bulkhead_test.go` |
| `implemented` | 当前 `UNKNOWN` 始终固定原连接；观察账本不等于健康度聚合 | 支付恢复与网关观察 | `TestSameProviderConnectionsStayPinnedThroughUnknownRecovery`、`TestPostgresProviderObservationLedgerFencesStaleGenerationAndIsImmutable` |

## 图表与配图

- 4 个 Mermaid：单次超时的双重事实、健康度状态、共享探测、异步观察流水线。
- 本章未添加 Ian 插图；状态机、时序和故障域表承担主要解释。
- 5 个 Go 代码块由正文提取，第 13 章测试覆盖朴素熔断器、分类、窗口、状态与舱壁隔离并发。

## 发布验证

- 书稿侧执行章节 Go 编译、竞态检测、静态检查、Mermaid 实际渲染、本地链接与 `git diff --check`。
- 路由服务 v3 与网关证据的代码证据以固定持续集成任务为准；健康度教学模型通过不等于生产健康度子系统已经存在。

## 未实现与后续章节

- 无 `services/router/internal/health`、健康度 PostgreSQL 迁移与摄取、快照发布器、共享探测租约，也没有网关按连接与资金操作隔离的舱壁。
- 无真实 PSP 健康度数据、仪表盘、自动事故覆盖或局部故障 Compose 端到端测试。
