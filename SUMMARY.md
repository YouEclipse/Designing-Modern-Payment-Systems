# 目录 / Table of Contents

* [中文导读](README.md)
* [配套 Go 微服务实现](https://github.com/YouEclipse/payments)

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

## English Edition

* [Introduction](en/README.md)

## Part I: Payment Foundations

* [Chapter 1: The Lifecycle of a Payment](en/part-1-foundations/payment-lifecycle.md)
* [Chapter 2: Why Payment Routing?](en/part-1-foundations/why-payment-routing.md)
* [Chapter 3: Responsibilities and Boundaries of a Payment Routing System](en/part-1-foundations/routing-system-boundaries.md)

## Part II: Building the Routing Engine

* [Chapter 4: Building Routing Context](en/part-2-routing-engine/routing-context.md)
* [Chapter 5: Rule-Based Routing](en/part-2-routing-engine/rule-based-routing.md)
* [Chapter 6: Provider Eligibility](en/part-2-routing-engine/provider-eligibility.md)
* [Chapter 7: Cost-Aware Routing](en/part-2-routing-engine/cost-aware-routing.md)
* [Chapter 8: Success-Rate Routing](en/part-2-routing-engine/success-rate-routing.md)
* [Chapter 9: Provider Scoring and Risk-Aware Routing](en/part-2-routing-engine/provider-scoring.md)

## Part III: Reliability

* [Chapter 10: Timeouts, Retries, and Unknown States](en/part-3-reliability/timeout-and-retry.md)
* [Chapter 11: Idempotency and Payment Attempts](en/part-3-reliability/idempotency.md)
* [Chapter 12: Fallback Routing](en/part-3-reliability/fallback-routing.md)
* [Chapter 13: Circuit Breakers and Provider Health](en/part-3-reliability/provider-health.md)

## Part IV: Operating at Scale

* [Chapter 14: Routing Decision Storage](en/part-4-operating-at-scale/routing-decision-store.md)
* [Chapter 15: Routing Observability](en/part-4-operating-at-scale/observability.md)
* [Chapter 16: Routing Experiments](en/part-4-operating-at-scale/experimentation.md)
* [Chapter 17: Reconciliation](en/part-4-operating-at-scale/reconciliation.md)
* [Chapter 18: Settlement and Cost Attribution](en/part-4-operating-at-scale/settlement.md)

## Part V: Global Routing and Evolution

* [Chapter 19: Multi-Region Payment Routing](en/part-5-global-routing/multi-region.md)
* [Chapter 20: The Routing Configuration Platform](en/part-5-global-routing/configuration-platform.md)
* [Chapter 21: Intelligent Routing](en/part-5-global-routing/intelligent-routing.md)

## Part VI: Integrated Case Study

* [Chapter 22: Designing a Global Payment Routing Platform from Scratch](en/part-6-case-study/global-payment-routing-platform.md)

## Appendices

* [Appendix A: State Models for Payment Intents, Attempts, and Provider Requests](en/appendices/payment-state-models.md)
* [Appendix B: Security, Compliance, and Data Boundaries](en/appendices/security-compliance.md)
* [Appendix C: Capacity Planning and Launch Checklist](en/appendices/capacity-planning-and-launch-checklist.md)
* [Glossary](en/glossary.md)
