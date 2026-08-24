# 附录 B 安全、合规与数据边界

> 状态：附录骨架

## 附录定位

支付路由天然接触 Merchant、交易和 Payment Method 元数据，但 Router 不应因此成为持有全部敏感数据的系统。本附录以 Data Minimization 和 Trust Boundary 为主线，整理设计 Routing Context、日志、配置和多区域架构时必须持续检查的安全与合规约束。

## 本附录将涵盖

* PCI DSS 作用域、PAN Tokenization、CVV 不落盘与敏感认证数据边界。
* 传输/静态加密、Key Management、Secret Rotation 与 Provider Credential 隔离。
* Service Identity、最小权限、RBAC、双人审批和 Break-Glass 审计。
* Decision Log、Trace、Metric 与事件流中的脱敏和保留周期。
* Data Residency、隐私删除、跨区域复制与目的限制。
* 配置投毒、重放、Provider Webhook 伪造和内部滥用的 Threat Model。
* 合规要求如何进入 Eligibility，而不把 Router 变成合规规则百科。

## 与正文的关系

第 4、14 章分别讨论 Context 与 Decision Record 的数据最小化，第 19、20 章讨论 Data Residency 和配置权限。本附录提供横向检查表；具体法律适用性仍需由目标市场的合规与法律团队确认。
