# 第 16 章 Routing Experimentation

> 状态：章节骨架

## 章节定位

离线回放无法完全预测新策略上线后的反馈效应，但直接随机切流又可能影响真实资金。本章设计从 Shadow Routing 到小流量 A/B Test、Champion/Challenger 的渐进式实验体系。

## 本章将涵盖

* Experiment Hypothesis、Primary Metric 与 Guardrail Metric。
* 按 Merchant、Customer 或 Payment Intent 的确定性分桶。
* Traffic Split、Shadow Decision、Replay 与 Champion/Challenger。
* Sample Ratio Mismatch、污染、网络效应和中途改配置。
* 成功率、成本、欺诈与拒付等延迟结果的观察窗口。
* 自动停止、人工 Kill Switch、回滚和实验审计。

## 推导路线

本章从 90/10 随机切流开始，通过同一 Intent 多次尝试跨组、Provider 容量变化和延迟拒付反馈等问题，加上稳定分桶、版本冻结和风险护栏，形成可逐级放量的实验流程。

## 与前后章节的关系

第 15 章提供实验测量基础，本章解决策略演进的因果评估。下一章面对另一类反馈：内部状态与 Provider 事实不一致时，系统必须通过 Reconciliation 恢复正确性，不能把观测结果直接当作最终真相。
