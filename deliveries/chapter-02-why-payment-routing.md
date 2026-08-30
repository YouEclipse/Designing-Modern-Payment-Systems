# 第 2 章《为什么需要支付路由》交付记录

## 章节契约

- 核心命题：单一 PSP 并非没有路由，而是路由决策被常量化；业务跨地区扩张后，覆盖硬约束、成本、表现、风险和故障不确定性迫使常量演进为可审计的决策流水线。
- 新增概念：多 PSP 动因、准入筛选、静态与规则路由、评分维度、路由决策、支付尝试、`UNKNOWN` 与后备路由边界。
- 留给后续章节：完整规则引擎、动态成功率与成本模型、风险评分、跨 PSP 后备路由、手动请款、退款、对账和真实 PSP 连接器。

## 能力状态

- `implemented`：固定实现中的静态准入决策、支付服务所有的执行、`UNKNOWN` 原操作恢复与单一资金路径。
- `teaching_model`：正文 2 个 Go 代码块中的硬编码反例与最小可解释路由服务。
- `target_contract`：本章尚未交付的跨支付服务商后备执行器、周期对账与真实服务商连接器；规则、成本、表现和风险的当前状态分别以第 5、7～9 章交付记录为准。

## 配套实现

- 代码仓库：[YouEclipse/payments](https://github.com/YouEclipse/payments)
- 实现提交：[bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)
- 持续集成任务：[33041147119 — 成功](https://github.com/YouEclipse/payments/actions/runs/33041147119)
- 纵向切片：令牌化银行卡自动请款，创建 → 确认 → 路由 → 支付服务商 → 回调通知与查询 → 持久收敛。

## 主张—实现—验证映射

| 状态 | 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- | --- |
| `implemented` | 准入筛选先于策略排序，并保留排除原因 | `services/router/internal/routing/engine.go` | `services/router/internal/routing/engine_test.go` |
| `implemented` | 路由服务只做决策，支付编排器执行资金副作用 | `services/router/`、`services/payment/internal/app/service.go` | 路由接口处理器测试与支付端到端测试 |
| `implemented` | 支付服务商超时属于 `UNKNOWN`，不盲切 PSP | `services/payment/internal/domain/model.go`、`services/payment/internal/app/reconciler.go` | `TestPaymentVerticalSlice/lost_response_becomes_unknown_and_converges_without_a_new_attempt` |
| `implemented` | 同一逻辑操作沿用原资金操作与服务商幂等键 | `services/provider-gateway/internal/gateway/service.go` | 网关崩溃恢复与请求哈希测试 |
| `implemented` | 重复或乱序事件不能制造第二条资金路径 | 支付收件箱、发件箱与网关回调通知存储 | PostgreSQL 集成测试；端到端测试断言只有一个支付尝试、一项支付资金操作和一项网关资金操作 |

## 图表与配图

- 4 张 Mermaid：单一服务商、多服务商、决策流水线与 `UNKNOWN` 时序，已实际解析。
- 5 张 Ian 风格 16:9 正文插图，位于 `assets/why-payment-routing-illustrations/`，均为 1672×941。

## 发布验证

- 中文正文约 6,900 汉字，符合样章 5,000～8,000 中文字目标。
- Go 示例、术语表、本地 Markdown 链接与配图路径通过检查。
- Go 单元测试、竞态检测、静态检查、OpenAPI、真实 PostgreSQL 集成测试、四镜像构建与 Compose 端到端测试全部通过。
- 代码与书稿远端提交值均已核对；模拟器明确限定为本地与测试用途，不宣称真实生产支付能力。

## 未实现与后续章节

- 固定提交仍保留本章用于教学起点的 v1 准入筛选与版本化静态策略；同一提交的 v2 与 v3 已实现规则、成本、观察性表现与风险护栏，窄边界和证据分别见第 5、7～9 章交付记录。它们不等于跨支付服务商后备执行器、因果实验或生产级智能路由。
- 当前支付切片只支持令牌化银行卡自动请款；真实 PSP、手动请款、撤销、退款、3DS、账务与清结算仍未实现。
