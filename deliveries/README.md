# 交付证据索引

本页汇总全书 22 章与 3 篇附录的交付清单。正文负责解释系统设计；交付清单负责固定配套实现提交、绿色持续集成、章节主张到代码与测试的映射、书稿验证与能力边界。它们不是真实资金生产认证，也不能把 `teaching_model` 或 `target_contract` 推导成已经存在的运行能力。

能力状态统一按以下口径读取：

- `implemented`：固定 `payments` 提交中存在，并有相应测试与绿色持续集成证据的窄能力；
- `teaching_model`：从正文提取、可编译并通过反例测试的 Go 教学模型；
- `target_contract`：未来实现必须满足的合同，没有运行能力已经完成的含义。

同一章可以同时包含三种状态。具体主张始终以对应交付记录的逐项分类和能力边界为准。

## 章节

| 章节 | 正文 | 交付清单 |
| --- | --- | --- |
| 第 1 章 | [一笔支付是如何发生的](../part-1-foundations/payment-lifecycle.md) | [交付记录](chapter-01-payment-lifecycle.md) |
| 第 2 章 | [为什么需要支付路由](../part-1-foundations/why-payment-routing.md) | [交付记录](chapter-02-why-payment-routing.md) |
| 第 3 章 | [支付路由系统的职责与边界](../part-1-foundations/routing-system-boundaries.md) | [交付记录](chapter-03-routing-system-boundaries.md) |
| 第 4 章 | [构建路由上下文](../part-2-routing-engine/routing-context.md) | [交付记录](chapter-04-routing-context.md) |
| 第 5 章 | [基于规则的路由](../part-2-routing-engine/rule-based-routing.md) | [交付记录](chapter-05-rule-based-routing.md) |
| 第 6 章 | [服务商准入筛选](../part-2-routing-engine/provider-eligibility.md) | [交付记录](chapter-06-provider-eligibility.md) |
| 第 7 章 | [成本感知路由](../part-2-routing-engine/cost-aware-routing.md) | [交付记录](chapter-07-cost-aware-routing.md) |
| 第 8 章 | [成功率路由](../part-2-routing-engine/success-rate-routing.md) | [交付记录](chapter-08-success-rate-routing.md) |
| 第 9 章 | [服务商评分与风险感知路由](../part-2-routing-engine/provider-scoring.md) | [交付记录](chapter-09-provider-scoring.md) |
| 第 10 章 | [超时、重试与未知状态](../part-3-reliability/timeout-and-retry.md) | [交付记录](chapter-10-timeout-and-retry.md) |
| 第 11 章 | [幂等性与支付尝试](../part-3-reliability/idempotency.md) | [交付记录](chapter-11-idempotency.md) |
| 第 12 章 | [后备路由](../part-3-reliability/fallback-routing.md) | [交付记录](chapter-12-fallback-routing.md) |
| 第 13 章 | [熔断器与服务商健康度](../part-3-reliability/provider-health.md) | [交付记录](chapter-13-provider-health.md) |
| 第 14 章 | [路由决策存储](../part-4-operating-at-scale/routing-decision-store.md) | [交付记录](chapter-14-routing-decision-store.md) |
| 第 15 章 | [路由可观测性](../part-4-operating-at-scale/observability.md) | [交付记录](chapter-15-routing-observability.md) |
| 第 16 章 | [路由实验](../part-4-operating-at-scale/experimentation.md) | [交付记录](chapter-16-routing-experimentation.md) |
| 第 17 章 | [对账](../part-4-operating-at-scale/reconciliation.md) | [交付记录](chapter-17-reconciliation.md) |
| 第 18 章 | [结算与成本归因](../part-4-operating-at-scale/settlement.md) | [交付记录](chapter-18-settlement-and-cost-attribution.md) |
| 第 19 章 | [多区域支付路由](../part-5-global-routing/multi-region.md) | [交付记录](chapter-19-multi-region-payment-routing.md) |
| 第 20 章 | [路由配置平台](../part-5-global-routing/configuration-platform.md) | [交付记录](chapter-20-routing-configuration-platform.md) |
| 第 21 章 | [智能路由](../part-5-global-routing/intelligent-routing.md) | [交付记录](chapter-21-intelligent-routing.md) |
| 第 22 章 | [从零设计全球支付路由平台](../part-6-case-study/global-payment-routing-platform.md) | [交付记录](chapter-22-global-payment-routing-platform.md) |

## 附录

| 附录 | 正文 | 交付清单 |
| --- | --- | --- |
| 附录 A | [支付意图、支付尝试与服务商请求状态模型](../appendices/payment-state-models.md) | [交付记录](appendix-a-payment-state-models.md) |
| 附录 B | [安全、合规与数据边界](../appendices/security-compliance.md) | [交付记录](appendix-b-security-compliance.md) |
| 附录 C | [容量规划与上线检查清单](../appendices/capacity-planning-and-launch-checklist.md) | [交付记录](appendix-c-capacity-planning-and-launch-checklist.md) |

全书目标合同与跨章所有权另见[《全书交付合同》](BOOK_CONTRACTS.md)。该合同描述应交付什么；当前实现状态仍以本页链接的逐章、逐附录交付记录为准。
