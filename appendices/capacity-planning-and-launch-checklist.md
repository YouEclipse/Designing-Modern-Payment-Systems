# 附录 C 容量规划与上线检查清单

> 状态：附录骨架

## 附录定位

Payment Router 位于同步支付关键路径，容量不足或上线准备遗漏会直接影响转化和资金安全。本附录把综合案例中的估算方法与各章的生产约束整理成可复用清单，但不会用一个固定 TPS 数字代替实际压测和故障演练。

## 本附录将涵盖

* 由峰值 TPS、Provider Fan-Out、Retry/Fallback 比例推导下游请求量。
* Routing Decision 的 CPU、缓存、Feature Lookup、存储和消息吞吐预算。
* Provider 限流、连接池、队列积压、热点 Merchant 与多 Region 余量。
* 正常流量、Provider 慢响应、缓存失效和区域切换下的容量测试。
* Rule/Capability/Cost 配置、Idempotency、Decision Store 与告警检查项。
* Provider 凭证、webhook、Reconciliation、Kill Switch 和回滚演练。
* 分阶段放量、Go/No-Go 标准、值班 Runbook 与上线后观测窗口。

## 与正文的关系

第 9 章定义在线决策延迟预算，第 13、15 章提供故障隔离和 SLO，第 19、20 章处理 Multi-Region 与配置发布，第 22 章给出具体规模案例。本附录用于实施核对，容量结论仍须使用目标架构、真实数据分布和 Provider 合同限制验证。
