# 第 5 章 Rule-Based Routing

> 状态：章节骨架

## 章节定位

当国家、币种、商户和 Provider 增多后，硬编码的 `if / else` 无法安全维护。本章将 Routing Policy 从应用代码中抽离，构建一个可验证、可版本化、可回滚且能够解释命中过程的 Rule Engine。

## 本章将涵盖

* Condition、Action、Priority、Scope 与 Effective Time 的 Rule Model。
* force、prefer、exclude、split 等动作及其适用边界。
* 多条规则同时命中时的优先级、合并与冲突处理。
* 不可变 Rule Set 版本、已发布快照与运行时版本选择协议。
* Rule Artifact 内部的 Schema 校验、语义校验、冲突检测，以及供外部回放使用的求值契约。
* 确定性求值、性能预算和规则命中解释。

## 推导路线

本章会从少量硬编码条件开始，让重复条件、跨商户例外和紧急改路逐步暴露问题；然后引入声明式规则，但继续追问规则冲突、无效 Provider 和版本漂移等生产风险，最终形成 Rule Artifact 与确定性求值器的契约。审批、灰度发布、跨 Region 分发和组织治理留到第 20 章的 Configuration Platform 展开。

## 与前后章节的关系

上一章定义 Routing Context，本章用其中的稳定字段表达确定性 Policy。下一章会补上 Rule 无法取代的硬约束过滤：先根据 Provider 能力、合同和合规要求得到合法 Candidate，再允许规则表达偏好或覆盖范围。
