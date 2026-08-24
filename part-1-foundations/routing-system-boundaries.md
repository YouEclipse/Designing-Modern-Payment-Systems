# 第 3 章 Payment Routing System 的职责与边界

> 状态：章节骨架

## 章节定位

第 2 章证明了 Routing Layer 的必要性，本章进一步回答它应该负责到哪里为止。边界如果过窄，决策缺少上下文且无法闭环；边界如果过宽，Router 会逐渐吞并支付状态机、Provider 调用和账务职责，成为难以演进的中心化单体。

## 本章将涵盖

* Payment API、Payment Orchestrator、Payment Router 与 Provider Adapter 的职责划分。
* Router 的输入、输出，以及“选择 Provider”和“执行支付”的差异。
* Adapter 如何归一化 authorize、capture、void、refund、query 与错误语义。
* Provider、Merchant Account、Connector、Region 与 Route Candidate 的区别。
* Routing Policy、Payment State 与 Provider State 的所有权。
* Router 应当保持无状态，还是需要持久化 Routing Decision。
* 在线决策路径与配置、指标、实验等控制平面的边界。
* 同步调用、异步事件与失败恢复中的责任归属。

## 推导路线

本章从一个同时完成编排、路由和 Provider 调用的简单服务开始，通过版本耦合、重试责任不清和审计困难等问题，推导出数据平面与控制平面、决策与执行相分离的生产架构。Provider Adapter 会被定义为协议与结果语义的隔离层，Router 只面向标准 Candidate 契约；最终给出组件契约，而不是宣称唯一正确的微服务拆分方式。

## 与前后章节的关系

前一章解释“为什么需要 Router”，本章定义“Router 是什么”。下一章将在这条边界内设计 Routing Context，明确一次决策究竟需要哪些输入，以及哪些敏感数据不应进入 Router。
