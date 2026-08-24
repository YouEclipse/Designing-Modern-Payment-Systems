# 附录 A Payment Intent、Attempt 与 Provider Request 状态模型

> 状态：附录骨架

## 附录定位

状态模型贯穿 Timeout、Idempotency、Fallback、webhook 和 Reconciliation。为避免各章重复定义并产生矛盾，本附录集中给出参考状态机、转换条件和关键不变量；正文仍会在具体问题出现时解释为什么需要这些状态。

## 本附录将涵盖

* Payment Intent、Payment Attempt、Provider Request 三层对象的职责和基数关系。
* `requires_action / processing / succeeded / failed / canceled` 等 Intent 状态。
* Attempt 的 `created / routed / executing / unknown / terminal` 状态与转换。
* Authorization、Capture、Void、Refund 等 Provider Operation 的独立状态。
* Approved、Declined、Technical Failure、Unknown Result 等 Outcome 分类及其恢复动作。
* webhook、Query、Retry、Fallback 和 Reconciliation 触发的转换优先级。
* 单调转换、终态保护、并发版本、事件去重和审计不变量。

## 与正文的关系

第 1 章首次介绍支付生命周期，第 10 至 12 章用这些对象处理 Unknown State 与多次尝试，第 14、17 章再把状态关联到 Decision Store 和 Reconciliation。本附录是参考模型，不暗示所有 Payment Method 都应共享一张状态表。
