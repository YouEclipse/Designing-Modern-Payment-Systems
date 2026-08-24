# 第 13 章 Circuit Breaker 与 Provider Health

> 状态：章节骨架

## 章节定位

当 Provider 出现区域性、方式级或商户账号级故障时，逐笔等待 Timeout 再 Fallback 会消耗延迟预算并形成级联故障。本章构建有维度、有时效的 Provider Health Model，并将其安全接入 Eligibility 与 Scoring。

## 本章将涵盖

* 主动探测与真实流量产生的被动健康信号。
* latency、timeout rate、technical error rate 与 authorization rate 的语义差异。
* `Healthy / Degraded / Unavailable / Recovering` 状态机。
* Circuit Breaker、Half-Open Probe、Bulkhead 与并发限制。
* 按 region、payment method、merchant account 等维度隔离故障。
* 流量切换后的观测偏差、恢复节流与 Thundering Herd。

## 推导路线

本章先使用单一全局布尔健康标记，再通过局部故障、低流量误判和恢复瞬间流量洪峰说明其不足，最终形成多维健康快照、状态迟滞和受控探测结合的设计。

## 与前后章节的关系

本章收束可靠性部分，并把运行时 Health 反馈给第 6、9 章的决策流水线。下一部分将从“系统当下如何决定”转向“团队事后如何解释、监控、实验并校正这些决定”。
