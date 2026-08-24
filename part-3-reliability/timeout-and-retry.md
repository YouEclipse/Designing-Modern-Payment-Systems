# 第 10 章 Timeout、Retry 与 Unknown State

> 状态：章节骨架

## 章节定位

Provider 调用超时只说明调用方没有及时收到结果，不说明支付没有发生。本章以“Issuer 已批准，但响应在网络中丢失”为核心故障，区分确定失败与 Unknown State，并建立支付场景下的 Timeout、Query 和 Retry 决策框架。

## 本章将涵盖

* 连接失败、读取超时、业务拒绝与 Provider 5xx 的不同语义。
* 端到端 Deadline、分层 Timeout 和剩余时间预算。
* `success / failure / unknown` 结果模型，以及 webhook 迟到与乱序。
* 何时可以 Retry，何时应先 Query，何时只能等待异步确认。
* Exponential Backoff、Jitter、Retry Budget 与 Retry Storm。
* 数据库或消息系统故障导致的本地状态不确定性。

## 推导路线

本章先对一次超时立即重试，再展示双重扣款如何产生；随后给请求增加状态查询和有限等待，但继续处理 Provider 不支持查询、查询也超时等情况，最终形成以结果确定性和副作用阶段为依据的恢复矩阵。

## 与前后章节的关系

上一章完成了 Provider 选择，本章说明执行结果可能不确定。下一章将引入稳定的 Payment Intent、Attempt 与 Provider Request 身份，使重放请求、并发调用和恢复流程能够识别“同一件事”。
