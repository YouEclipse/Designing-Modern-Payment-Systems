# 第 17 章 Reconciliation

> 状态：章节骨架

## 章节定位

即使在线链路设计正确，Payment Platform 与 Provider 仍会因 Timeout、迟到 webhook、人工处理或文件延迟而产生状态差异。本章把 Reconciliation 设计为独立的事实比对和受控修复流程，而不是一条临时 SQL。

## 本章将涵盖

* API 响应、webhook、Provider Query 与结算文件的事实层级。
* Provider Transaction ID、Merchant Reference 与金额币种的 Matching。
* 差异类型：Missing、Duplicate、Status Mismatch、Amount Mismatch。
* 增量扫描、窗口回看、Checkpoint 与幂等任务执行。
* 自动修复、人工调查、Ledger Adjustment 与审计证据。
* Reconciliation 结果如何校正成功率和成本数据。

## 推导路线

本章会先信任 webhook 更新状态，再引入丢失、重复、乱序和 Provider 侧人工变更，逐步形成多源摄取、匹配、差异分类和修复工作流，并明确“发现差异”与“改写资金事实”的权限边界。

## 与前后章节的关系

上一章讨论如何验证新的 Routing Policy，本章保证评价所依据的最终结果可信。下一章把修正后的 Provider 与结算数据关联回 Routing Decision，计算实际费用、FX 和每笔成功支付成本。
