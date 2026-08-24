# 第 11 章 Idempotency 与支付尝试

> 状态：章节骨架

## 章节定位

Idempotency 不是在 Redis 中缓存一个响应，而是定义重复请求的业务身份、并发所有权和可重放结果。本章围绕 Payment Intent、Payment Attempt 与 Provider Request 三层对象，设计端到端去重机制。

## 本章将涵盖

* Idempotency Key 的作用域、生命周期、参数指纹和冲突语义。
* Payment Intent、Attempt、Routing Decision、Provider Request 的一对多关系。
* 数据库唯一约束、原子插入与并发请求的胜者判定。
* 处理中请求的响应、结果持久化与确定性 Replay。
* Provider 原生 Idempotency Key 的映射和不支持时的补偿策略。
* API、队列消费者、webhook 与定时恢复任务的幂等边界。

## 推导路线

本章会从内存缓存去重开始，通过进程重启、并发竞态和相同 Key 不同参数等反例逐步加固设计；最终以持久化业务对象和唯一约束作为事实来源，缓存只承担加速职责。

## 与前后章节的关系

第 10 章建立了 Unknown State 问题，本章提供安全恢复所需的身份模型。下一章会利用 Attempt 边界判断能否切换 Provider，并说明幂等只能阻止同一请求重复执行，不能自动证明跨 Provider Fallback 安全。
