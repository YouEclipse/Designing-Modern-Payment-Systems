# 第 21 章 Intelligent Routing

> 状态：章节骨架

## 章节定位

当规则和人工权重无法适应快速变化、细粒度的 Provider 表现时，统计学习可能提高决策质量。本章从现有 Scoring Pipeline 演进到 Contextual Bandit 与 ML Scoring，同时认真处理支付场景中的探索成本、偏差和延迟反馈。

## 本章将涵盖

* 何时规则足够，何时问题具备使用 ML 的数据与收益条件。
* Supervised Scoring、Contextual Bandit 与 Reinforcement Learning 的边界。
* Logged Policy Bias、Counterfactual Evaluation 与 Selection Bias。
* Sparse Data、Cold Start、Non-Stationarity 和 Delayed Feedback。
* Feature Pipeline、Training、Model Registry、Online Serving 与版本快照。
* Eligibility/Policy Guardrail、安全探索、Fallback Model 和快速降级。

## 推导路线

本章先让模型预测每个 Provider 的成功概率，再指出训练数据只来自历史已选路线、失败与拒付反馈延迟且环境持续变化。随后加入离线评估、受控探索和规则护栏，最后给出“保持 Rule-Based”也可能是正确结论的决策清单。

## 与前后章节的关系

前两章提供全球执行与安全发布基础，本章是路由策略的可选演进，而非全书默认终点。下一章综合案例会同时给出规则优先的初始版本和可插拔优化路径，验证此前组件能否组成完整系统。
