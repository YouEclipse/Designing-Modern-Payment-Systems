# 第 14 章 Routing Decision Store

> 状态：章节骨架

## 章节定位

生产系统必须能回答“这笔支付为什么去了 PSP B”，而不是只记录最终 Provider。本章设计 Append-Only 的 Routing Decision Record，使一次决策可以被审计、复现、调试，并作为实验与成本归因的数据基线。

## 本章将涵盖

* Decision ID 与 Payment Intent、Attempt、Provider Request 的关联。
* Context 摘要、候选列表、排除原因、规则命中和分数分解。
* Rule、Feature、Health Snapshot、Experiment 等版本引用。
* Selected Provider、Fallback Plan、Reason Code 与决策时延。
* 不可变记录、Schema 演进、冷热分层和保留周期。
* PCI/隐私数据最小化、访问控制与审计日志。

## 推导路线

本章从一条 `selected_provider=PSP_B` 日志开始，使用配置已变更、指标已滚动和多次 Fallback 的排障场景逐步补齐证据。最终记录足够的决策事实和版本引用，但不复制完整敏感上下文或把 Store 放到选择路径的强依赖上。

## 与前后章节的关系

第 13 章完成在线可靠性闭环，本章把在线决策变成可查询事实。下一章将在这些事实之上定义指标、追踪和告警，同时区分 Payment Outcome、Provider Performance 与 Router 自身健康。
