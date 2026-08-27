# 交付证据索引

本页汇总全书 22 章与 3 篇附录的 delivery manifest。正文负责解释系统设计；manifest 负责固定配套实现 commit、绿色 CI、章节主张到代码/测试的映射、书稿验证与能力边界。它们不是 live-payment 认证，也不能把 `teaching_model` 或 `target_contract` 推导成已经存在的 runtime。

能力状态统一按以下口径读取：

- `implemented`：固定 `payments` commit 中存在，并有相应测试与绿色 CI 证据的窄能力；
- `teaching_model`：从正文提取、可编译并通过反例测试的 Go 教学模型；
- `target_contract`：未来实现必须满足的合同，没有 runtime 完成含义。

同一章可以同时包含三种状态。具体主张始终以对应 manifest 的逐项分类和能力边界为准。

## 章节

| 章节 | 正文 | Delivery manifest |
| --- | --- | --- |
| 第 1 章 | [一笔支付是如何发生的](../part-1-foundations/payment-lifecycle.md) | [交付记录](chapter-01-payment-lifecycle.md) |
| 第 2 章 | [为什么需要 Payment Routing](../part-1-foundations/why-payment-routing.md) | [交付记录](chapter-02-why-payment-routing.md) |
| 第 3 章 | [Payment Routing System 的职责与边界](../part-1-foundations/routing-system-boundaries.md) | [交付记录](chapter-03-routing-system-boundaries.md) |
| 第 4 章 | [构建 Routing Context](../part-2-routing-engine/routing-context.md) | [交付记录](chapter-04-routing-context.md) |
| 第 5 章 | [Rule-Based Routing](../part-2-routing-engine/rule-based-routing.md) | [交付记录](chapter-05-rule-based-routing.md) |
| 第 6 章 | [Provider Eligibility](../part-2-routing-engine/provider-eligibility.md) | [交付记录](chapter-06-provider-eligibility.md) |
| 第 7 章 | [Cost-Aware Routing](../part-2-routing-engine/cost-aware-routing.md) | [交付记录](chapter-07-cost-aware-routing.md) |
| 第 8 章 | [Success-Rate Routing](../part-2-routing-engine/success-rate-routing.md) | [交付记录](chapter-08-success-rate-routing.md) |
| 第 9 章 | [Provider Scoring 与 Risk-Aware Routing](../part-2-routing-engine/provider-scoring.md) | [交付记录](chapter-09-provider-scoring.md) |
| 第 10 章 | [Timeout、Retry 与 Unknown State](../part-3-reliability/timeout-and-retry.md) | [交付记录](chapter-10-timeout-and-retry.md) |
| 第 11 章 | [Idempotency 与支付尝试](../part-3-reliability/idempotency.md) | [交付记录](chapter-11-idempotency.md) |
| 第 12 章 | [Fallback Routing](../part-3-reliability/fallback-routing.md) | [交付记录](chapter-12-fallback-routing.md) |
| 第 13 章 | [Circuit Breaker 与 Provider Health](../part-3-reliability/provider-health.md) | [交付记录](chapter-13-provider-health.md) |
| 第 14 章 | [Routing Decision Store](../part-4-operating-at-scale/routing-decision-store.md) | [交付记录](chapter-14-routing-decision-store.md) |
| 第 15 章 | [Routing Observability](../part-4-operating-at-scale/observability.md) | [交付记录](chapter-15-routing-observability.md) |
| 第 16 章 | [Routing Experimentation](../part-4-operating-at-scale/experimentation.md) | [交付记录](chapter-16-routing-experimentation.md) |
| 第 17 章 | [Reconciliation](../part-4-operating-at-scale/reconciliation.md) | [交付记录](chapter-17-reconciliation.md) |
| 第 18 章 | [Settlement 与 Cost Attribution](../part-4-operating-at-scale/settlement.md) | [交付记录](chapter-18-settlement-and-cost-attribution.md) |
| 第 19 章 | [Multi-Region Payment Routing](../part-5-global-routing/multi-region.md) | [交付记录](chapter-19-multi-region-payment-routing.md) |
| 第 20 章 | [Routing Configuration Platform](../part-5-global-routing/configuration-platform.md) | [交付记录](chapter-20-routing-configuration-platform.md) |
| 第 21 章 | [Intelligent Routing](../part-5-global-routing/intelligent-routing.md) | [交付记录](chapter-21-intelligent-routing.md) |
| 第 22 章 | [从零设计 Global Payment Routing Platform](../part-6-case-study/global-payment-routing-platform.md) | [交付记录](chapter-22-global-payment-routing-platform.md) |

## 附录

| 附录 | 正文 | Delivery manifest |
| --- | --- | --- |
| 附录 A | [Payment Intent、Attempt 与 Provider Request 状态模型](../appendices/payment-state-models.md) | [交付记录](appendix-a-payment-state-models.md) |
| 附录 B | [安全、合规与数据边界](../appendices/security-compliance.md) | [交付记录](appendix-b-security-compliance.md) |
| 附录 C | [容量规划与上线检查清单](../appendices/capacity-planning-and-launch-checklist.md) | [交付记录](appendix-c-capacity-planning-and-launch-checklist.md) |

全书目标合同与跨章所有权另见[《全书交付合同》](BOOK_CONTRACTS.md)。该合同描述应交付什么；当前实现状态仍以本页链接的逐章、逐附录 manifest 为准。
