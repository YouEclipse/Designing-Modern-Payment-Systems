# 第 2 章《为什么需要 Payment Routing》交付记录

## 章节契约

- 核心命题：单 PSP 并非没有路由，而是路由决策被常量化；业务跨地区扩张后，覆盖硬约束、成本、表现、风险和故障不确定性迫使常量演进为可审计的决策流水线。
- 新增概念：多 PSP 动因、Eligibility、静态与规则路由、评分维度、Routing Decision、Payment Attempt、UNKNOWN 与 fallback 边界。
- 留给后续章节：完整 Rule Engine、动态成功率与成本模型、风险评分、跨 PSP fallback、manual capture、refund、对账和真实 PSP connector。

## 配套实现

- Repository：[YouEclipse/payments](https://github.com/YouEclipse/payments)
- Implementation commit：[b5de28105f5de92de1c1226fbed5c57442984906](https://github.com/YouEclipse/payments/commit/b5de28105f5de92de1c1226fbed5c57442984906)
- CI run：[32759532120](https://github.com/YouEclipse/payments/actions/runs/32759532120)
- 纵向切片：tokenized card automatic capture，create → confirm → route → Provider → webhook/query → durable convergence。

## 主张—实现—验证映射

| 章节主张或不变量 | 工程文件 | 测试或运行证据 |
| --- | --- | --- |
| Eligibility 先于策略排序，并保留排除原因 | `services/router/internal/routing/engine.go` | `services/router/internal/routing/engine_test.go` |
| Router 只做决策，Orchestrator 执行资金副作用 | `services/router/`、`services/payment/internal/app/service.go` | Router handler tests 与 payment E2E |
| Provider timeout 是 UNKNOWN，不盲切 PSP | `services/payment/internal/domain/model.go`、`services/payment/internal/app/reconciler.go` | `TestPaymentVerticalSlice/lost_response_becomes_unknown_and_converges_without_a_new_attempt` |
| 同一逻辑操作沿用原 operation 与 Provider key | `services/provider-gateway/internal/gateway/service.go` | Gateway crash-recovery 与 request-hash tests |
| 重复或乱序事件不能制造第二条资金路径 | payment inbox/outbox 与 gateway webhook store | PostgreSQL integration tests；E2E 断言一个 Attempt、一个 Payment Operation、一个 Gateway Operation |

## 图表与配图

- 4 张 Mermaid：单 Provider、多 Provider、决策流水线与 UNKNOWN 时序，已实际解析。
- 5 张 Ian 风格 16:9 正文插图，位于 `assets/why-payment-routing-illustrations/`，均为 1672×941。

## 发布验证

- 中文正文约 6,900 汉字，符合样章 5,000～8,000 中文字目标。
- Go 示例、术语表、本地 Markdown 链接与配图路径通过检查。
- Go unit、race、vet、OpenAPI、真实 PostgreSQL integration、四镜像构建与 Compose E2E 全绿。
- 代码与书稿远端 SHA 均已核对；Simulator 明确限定 local/test，不宣称 live-payment 能力。

## 未实现与后续章节

- 当前 Router 只有 Eligibility 与版本化静态策略，尚未实现完整规则求值、成本、成功率和风险评分。
- 当前支付切片只支持 tokenized card automatic capture；真实 PSP、manual capture、Void、Refund、3DS、账务与清结算仍待后续章节。
