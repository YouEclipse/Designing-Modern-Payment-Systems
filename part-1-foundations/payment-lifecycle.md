# 第 1 章 一笔支付是如何发生的

> 状态：章节骨架

## 章节定位

本章从一笔银行卡支付出发，为没有支付行业经验的读者建立最小但完整的领域模型。重点不是罗列机构名词，而是沿着请求与资金的真实路径，解释 Merchant、Payment Platform、PSP、Acquirer、Card Network 与 Issuer 如何协作，以及 Authorization 与资金真正转移之间为何存在时间差。

## 本章将涵盖

* 一次线上支付中 Customer、Merchant 与各支付机构的职责。
* Authorization、Capture、Void、Refund、Chargeback 的目的与先后关系。
* Payment Intent、Payment Attempt、Provider Transaction 等不同标识代表的对象。
* 同步 API 响应、异步 webhook 与后续 Settlement 之间的关系。
* 卡支付与钱包、银行转账等 Payment Method 在生命周期上的共同点与差异。
* “请求失败”不等于“支付失败”这一可靠性问题的最初伏笔。

## 推导路线

本章会先画出最短的支付链路，再逐步加入 Acquirer、Card Network 和 Issuer；随后用 Authorization 与 Capture 分离的场景说明，业务状态、Provider 状态和资金状态并不是同一个状态。章节结束时，读者应能读懂后续所有 Routing 示例中的交易上下文和结果状态。

## 与前后章节的关系

这是全书的领域入口，不要求读者预先了解支付。下一章会保持同一条支付链路，但把单一 PSP 替换为多个 Provider，由可用性、覆盖范围、成本和成功率问题自然推导出 Payment Routing。
