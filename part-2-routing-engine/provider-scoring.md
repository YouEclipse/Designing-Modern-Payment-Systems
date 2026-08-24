# 第 9 章 Provider Scoring 与 Risk-Aware Routing

> 状态：章节骨架

## 章节定位

当合法候选不止一个时，Router 需要把多种目标转化为可比较的决策。本章不会停留在一个 weighted sum 公式，而会完整设计从 Candidate Generation 到 Tie-Breaking、Selection 与 Decision Record 的在线流水线，并划清 Risk-Aware Routing 与 Fraud Engine 的边界。

## 本章将涵盖

* Eligibility、Policy、Feature Lookup、Scoring、Selection 的执行顺序。
* 不同量纲的归一化、权重、阈值和硬约束。
* Risk Decision、Risk Score、Reason Code 与信号时效的标准契约。
* 3DS、SCA、AVS、Token 等能力如何成为 Eligibility 或 Scoring 输入。
* 中央 Risk Block 不被 Fallback 绕过，以及 Risk Signal 缺失时的降级。
* Expected Margin、Cost per Success 与多目标优化的候选方案。
* 缺失或过期 Feature、单候选、同分和无候选时的降级。
* 决策确定性、稳定 Tie-Breaker 与流量抖动控制。
* Score Breakdown、Reason Code 与决策时延预算。

## 推导路线

本章先构建直观的加权评分，再用量纲变化、目标相互补偿和数据过期等反例检验它；随后加入 Risk Contract，明确风险系统决定允许什么，Router 只在允许集合内选择执行路径。最终区分不可违反的 Guardrail 与可权衡的 Objective，形成可插拔但执行顺序稳定的生产决策流水线。

## 与前后章节的关系

第 7、8 章分别提供成本与成功概率，本章把它们与 Policy、Risk 和 Health 组合并收束第二部分。下一章会挑战一个隐含前提：被选中的 Provider 调用并不总能返回确定结果，Timeout、Retry 与 Unknown State 会改变后续动作。
