# Table of contents

* [导读](README.md)
* [配套 Go 微服务实现](https://github.com/YouEclipse/payments)
* [交付证据索引](deliveries/README.md)

## 第一部分 支付基础

* [第 1 章 一笔支付是如何发生的](part-1-foundations/payment-lifecycle.md)
* [第 2 章 为什么需要 Payment Routing](part-1-foundations/why-payment-routing.md)
* [第 3 章 Payment Routing System 的职责与边界](part-1-foundations/routing-system-boundaries.md)

## 第二部分 构建 Routing Engine

* [第 4 章 构建 Routing Context](part-2-routing-engine/routing-context.md)
* [第 5 章 Rule-Based Routing](part-2-routing-engine/rule-based-routing.md)
* [第 6 章 Provider Eligibility](part-2-routing-engine/provider-eligibility.md)
* [第 7 章 Cost-Aware Routing](part-2-routing-engine/cost-aware-routing.md)
* [第 8 章 Success-Rate Routing](part-2-routing-engine/success-rate-routing.md)
* [第 9 章 Provider Scoring 与 Risk-Aware Routing](part-2-routing-engine/provider-scoring.md)

## 第三部分 可靠性

* [第 10 章 Timeout、Retry 与 Unknown State](part-3-reliability/timeout-and-retry.md)
* [第 11 章 Idempotency 与支付尝试](part-3-reliability/idempotency.md)
* [第 12 章 Fallback Routing](part-3-reliability/fallback-routing.md)
* [第 13 章 Circuit Breaker 与 Provider Health](part-3-reliability/provider-health.md)

## 第四部分 规模化运营

* [第 14 章 Routing Decision Store](part-4-operating-at-scale/routing-decision-store.md)
* [第 15 章 Routing Observability](part-4-operating-at-scale/observability.md)
* [第 16 章 Routing Experimentation](part-4-operating-at-scale/experimentation.md)
* [第 17 章 Reconciliation](part-4-operating-at-scale/reconciliation.md)
* [第 18 章 Settlement 与 Cost Attribution](part-4-operating-at-scale/settlement.md)

## 第五部分 全球化与路由演进

* [第 19 章 Multi-Region Payment Routing](part-5-global-routing/multi-region.md)
* [第 20 章 Routing Configuration Platform](part-5-global-routing/configuration-platform.md)
* [第 21 章 Intelligent Routing](part-5-global-routing/intelligent-routing.md)

## 第六部分 综合案例

* [第 22 章 从零设计 Global Payment Routing Platform](part-6-case-study/global-payment-routing-platform.md)

## 附录

* [附录 A Payment Intent、Attempt 与 Provider Request 状态模型](appendices/payment-state-models.md)
* [附录 B 安全、合规与数据边界](appendices/security-compliance.md)
* [附录 C 容量规划与上线检查清单](appendices/capacity-planning-and-launch-checklist.md)
* [术语表](glossary.md)
