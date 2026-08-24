# 第 12 章 Fallback Routing

> 状态：章节骨架

## 章节定位

Fallback 能提高可用性，也能把一次网络抖动放大为重复扣款。本章把 Fallback 定义为受约束的新 Payment Attempt，依据前一次结果是否确定、拒绝类型和 Payment Method 语义决定是否允许换路。

## 本章将涵盖

* Technical Failure、Soft Decline、Hard Decline 与 Unknown Result 的分类。
* Safe Fallback 和 Unsafe Fallback 的判定矩阵。
* 同 Provider Retry、Provider Query 与跨 Provider Fallback 的先后关系。
* Route Plan、已尝试集合、最大次数、总 Deadline 和循环防护。
* 3DS、银行转账、钱包等带用户交互方式的特殊限制。
* 原 Provider 迟到响应与 Fallback 成功同时发生时的冲突处理。

## 推导路线

本章先实现“PSP A 失败就调用 PSP B”，再分别放入明确拒绝、连接失败、读取超时和迟到 webhook，观察哪些场景会伤害成功率或制造重复支付。最终得到由 Outcome Classification 驱动的 Fallback Engine。

## 与前后章节的关系

上一章给每次尝试稳定身份，本章据此组织多次受控尝试。下一章把单笔调用的失败扩展为 Provider 的持续性健康问题，让新的交易在进入失败路径之前就绕开异常通道。
