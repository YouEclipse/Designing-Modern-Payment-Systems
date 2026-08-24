# 第 15 章 Routing Observability

> 状态：章节骨架

## 章节定位

Routing Observability 不只是看总体 Authorization Rate。本章建立从 Router 服务 SLI、Provider 调用质量到业务结果与成本的分层观测体系，避免聚合指标掩盖某个国家、Issuer 或 Payment Method 的回归。

## 本章将涵盖

* Routing Decision latency、error、no-candidate 与 stale-feature 指标。
* authorization、decline、timeout、fallback、routing distribution 等业务指标。
* Cost per Successful Payment 与结果延迟带来的口径问题。
* Trace 如何串联 Intent、Attempt、Decision 与 Provider Request。
* 高基数维度控制、预聚合、Sampling 与日志检索。
* SLO、分层告警、基线异常和可执行的 Runbook。

## 推导路线

本章先建立少量全局 Dashboard，再通过 Simpson's Paradox、流量结构变化和 Provider 延迟回执等案例，推导出按决策阶段与业务切片分层的指标模型，并明确每类告警的负责人和恢复动作。

## 与前后章节的关系

上一章提供可解释的数据源，本章把它变成持续反馈。下一章利用同一套指标和 Decision Record 评估新策略，但会额外处理实验分桶、选择偏差和支付业务的安全护栏。
