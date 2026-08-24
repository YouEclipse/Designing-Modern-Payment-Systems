# 第 7 章 Cost-Aware Routing

> 状态：章节骨架

## 章节定位

同一笔支付通过不同 Provider 的真实成本可能显著不同。本章建立可用于在线决策的 Expected Cost Model，并解释为什么“选择报价最低的 Provider”既可能算错成本，也可能降低最终利润。

## 本章将涵盖

* fixed fee、percentage fee、scheme fee、cross-border fee、FX markup 等成本项。
* 阶梯价、月度承诺量、refund fee 与 chargeback cost 的建模边界。
* 按交易估算 `expected_cost(transaction, provider)` 的数据与精度。
* 成本配置的币种换算、有效期、版本与账单校准。
* Cost per Attempt 与 Cost per Successful Payment 的区别。
* 成本、成功率、延迟和可靠性之间的权衡。

## 推导路线

本章先比较两家 PSP 的表面费率，再加入固定费、跨境费和失败尝试成本，展示局部最便宜为何不等于期望成本最低。随后设计一个可解释的成本分解，并保留估算值与结算后实际值的反馈链路。

## 与前后章节的关系

第 5、6 章分别处理显式 Policy 与硬约束，本章加入第一个连续优化目标。下一章会估计 Provider 对当前细分交易的成功概率；两者结合后，系统才能比较“每次尝试便宜”和“每笔成功支付便宜”。
