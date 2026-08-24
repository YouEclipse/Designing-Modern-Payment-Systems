# 第 18 章 Settlement 与 Cost Attribution

> 状态：章节骨架

## 章节定位

Routing Decision 的经济影响在 Authorization 成功时尚未结束。手续费、FX、退款、Reserve 和 Chargeback 可能在数日后出现。本章解释如何把结算事实归因到 Route 与 Attempt，使 Cost-Aware Routing 拥有可校准的真实反馈。

## 本章将涵盖

* Clearing、Settlement、Payout 与 Ledger 的关系和边界。
* Provider Fee、Scheme Fee、FX、Reserve、Refund 与 Chargeback 成本。
* Provider 报表到 Payment Attempt、Decision 与 Merchant 的关联键。
* Expected Cost、Accrued Cost 与 Actual Cost 的差异。
* 多次 Attempt、部分 Capture/Refund 和跨币种下的成本分摊。
* Fee Model 校准、异常账单检测和单位经济性指标。

## 推导路线

本章从授权时估算的一笔费率开始，沿结算文件逐项加入实际费用和后续事件，揭示为何简单按成功交易计费会扭曲 Provider 表现。最终形成从 Decision 到 Settlement Line Item 的可追溯归因链路。

## 与前后章节的关系

本章完成运营反馈闭环，并把实际成本送回第 7、9 章。下一部分转向全球部署与变更治理：当规则、数据与决策跨团队、跨区域传播时，如何保持安全、可用和可解释。
