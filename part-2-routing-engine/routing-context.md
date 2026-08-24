# 第 4 章 构建 Routing Context

> 状态：章节骨架

## 章节定位

Routing Decision 的质量受输入质量约束。本章把分散在订单、支付方式、商户配置、风险系统与运行时健康状态中的数据，整理为一个版本化、可审计的 Routing Context，同时避免让 Router 依赖每个上游系统的内部模型。

## 本章将涵盖

* Merchant、Transaction、Payment Method、Customer、Risk 与 Runtime Signal 的字段分组。
* amount 的最小货币单位、currency、country、card brand、BIN/IIN 等字段语义。
* 已知值、缺失值、派生值和实时值的来源及可信度。
* Context Schema 的版本、向后兼容、默认值与校验策略。
* PCI 数据最小化、Token 使用和敏感字段脱敏。
* 如何保存决策快照，同时避免记录不必要的个人数据。

## 推导路线

本章会先展示参数不断增长的 `route(...)` 函数，再分析跨服务查询、字段语义漂移和不可复现决策的问题，最终形成稳定的 Routing Context 契约，并明确哪些数据应在 Context Builder 中组装、哪些只能作为带时效的引用。

## 与前后章节的关系

第 3 章给出了 Router 的 API 边界，本章把输入具体化。下一章将从最简单的硬编码条件出发，把随 Merchant、国家和 Payment Method 增长的 Routing Policy 抽离为可验证的 Rule Model。
