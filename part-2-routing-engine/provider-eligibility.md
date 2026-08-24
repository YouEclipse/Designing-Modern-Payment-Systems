# 第 6 章 Provider Eligibility

> 状态：章节骨架

## 章节定位

在比较成本或成功率之前，系统必须先回答 Provider 是否有资格处理当前交易。Eligibility 是硬约束过滤，不是偏好评分；混淆二者可能让高分但不合规、无合同或不支持当前 Payment Method 的 Provider 被选中。

## 本章将涵盖

* 国家、币种、Payment Method、Card Brand、金额区间等能力约束。
* Merchant Contract、Provider Account、合规要求与区域限制。
* Provider Adapter 如何归一化 API 能力、Decline Code、幂等和 webhook 语义。
* Capability Catalog 的版本，以及静态能力与 Connector 配置的区别。
* 静态配置、动态 Health 和交易上下文如何共同产生 Candidate Set。
* `eligible / ineligible / unknown` 三值结果与原因码。
* 无候选、仅一个候选和候选数据过期时的处理策略。
* Eligibility 结果的缓存、失效与审计。

## 推导路线

本章先尝试对所有 Provider 直接评分，再通过“不支持该币种的 Provider 得分最高”等反例说明评分不能取代硬约束。随后建立由 Provider Adapter 暴露、按 Connector 和 Merchant Account 细分的 Capability Catalog；最终形成可解释的过滤流水线，并为每个被剔除的候选保留结构化原因。

## 与前后章节的关系

第 5 章把显式业务 Policy 从代码中抽离，本章确保这些 Policy 只能作用于合法 Candidate。下一章将在候选集合上加入 Expected Cost，开始处理能够权衡、但不能突破硬约束的优化目标。
