# 目录

* [导读](README.md)
* [配套 Go 微服务实现](https://github.com/YouEclipse/payments)
* [交付证据索引](deliveries/README.md)

## 第一部分 支付基础

* [第 1 章 一笔支付是如何发生的](part-1-foundations/payment-lifecycle.md)
* [第 2 章 为什么需要支付路由](part-1-foundations/why-payment-routing.md)
* [第 3 章 支付路由系统的职责与边界](part-1-foundations/routing-system-boundaries.md)

## 第二部分 构建路由引擎

* [第 4 章 构建路由上下文](part-2-routing-engine/routing-context.md)
* [第 5 章 基于规则的路由](part-2-routing-engine/rule-based-routing.md)
* [第 6 章 服务商准入筛选](part-2-routing-engine/provider-eligibility.md)
* [第 7 章 成本感知路由](part-2-routing-engine/cost-aware-routing.md)
* [第 8 章 成功率路由](part-2-routing-engine/success-rate-routing.md)
* [第 9 章 服务商评分与风险感知路由](part-2-routing-engine/provider-scoring.md)

## 第三部分 可靠性

* [第 10 章 超时、重试与未知状态](part-3-reliability/timeout-and-retry.md)
* [第 11 章 幂等性与支付尝试](part-3-reliability/idempotency.md)
* [第 12 章 后备路由](part-3-reliability/fallback-routing.md)
* [第 13 章 熔断器与服务商健康度](part-3-reliability/provider-health.md)

## 第四部分 规模化运营

* [第 14 章 路由决策存储](part-4-operating-at-scale/routing-decision-store.md)
* [第 15 章 路由可观测性](part-4-operating-at-scale/observability.md)
* [第 16 章 路由实验](part-4-operating-at-scale/experimentation.md)
* [第 17 章 对账](part-4-operating-at-scale/reconciliation.md)
* [第 18 章 结算与成本归因](part-4-operating-at-scale/settlement.md)

## 第五部分 全球化与路由演进

* [第 19 章 多区域支付路由](part-5-global-routing/multi-region.md)
* [第 20 章 路由配置平台](part-5-global-routing/configuration-platform.md)
* [第 21 章 智能路由](part-5-global-routing/intelligent-routing.md)

## 第六部分 综合案例

* [第 22 章 从零设计全球支付路由平台](part-6-case-study/global-payment-routing-platform.md)

## 附录

* [附录 A 支付意图、支付尝试与服务商请求状态模型](appendices/payment-state-models.md)
* [附录 B 安全、合规与数据边界](appendices/security-compliance.md)
* [附录 C 容量规划与上线检查清单](appendices/capacity-planning-and-launch-checklist.md)
* [术语表](glossary.md)
