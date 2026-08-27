# Designing Modern Payment Systems

## 设计现代支付系统：从一笔支付到全球支付路由平台

现代支付系统面对的核心问题，不只是“如何调用一家 PSP 的 API”，而是如何在多个国家、币种、支付方式、商户实体和 Provider 之间，持续做出安全、可解释、可运营的支付决策。

本书以 **支付路由系统（Payment Routing System）** 为主线，从最简单的单 PSP 架构出发。当业务逐步遇到覆盖差异、授权成功率、处理成本、风险约束、Provider 故障和全球部署等问题时，再推导出 Routing Context、规则引擎、Provider Eligibility、动态评分、fallback、决策审计与 reconciliation 等生产级能力。

这不是一本支付术语百科，也不会先给出一张复杂架构图再逐项解释组件。主要章节采用下面的演进方式：

```text
Problem → Simple Design → Problems at Scale → Production Design → Trade-offs
```

读者会先看到一个足够简单的系统如何工作，再观察它在哪些真实条件下失效，最后理解新组件为什么出现、解决了什么问题，又带来了哪些代价。

> 中文版正文第一版已经完整交付：全书包含 22 章、3 篇附录与术语表；正文中的 Go 教学示例、Mermaid 图和明确引用的当前实现均纳入可重复验证。配套系统只在各章 delivery manifest 标为 `implemented` 且由固定代码 SHA/绿色 CI 支撑的范围内算已实现；`teaching_model` 与 `target_contract` 不代表对应 runtime 已存在。英文版将在此基础上整体翻译并重写。

## 代码示例约定

本书所有用于展示后端实现的示例代码统一使用 **Go**，包括领域模型、API、支付状态机、路由规则求值、Provider Adapter、幂等与重试、数据访问以及并发处理。示例会优先使用清晰的类型和标准库，并尽可能保持可编译、可测试，而不是只给出无法运行的片段。

JSON、SQL、Mermaid 和少量伪代码仍会分别用于表达 API 数据、存储模型、架构时序和决策公式；它们只承担对应的描述职责，不作为另一套实现语言。

## 配套 Go 支付系统

本书同步维护一个可运行的微服务实现：[YouEclipse/payments](https://github.com/YouEclipse/payments)。它不是把章节里的片段简单拼在一起，而是用 Payment、Router、Provider Gateway 和仅限开发/测试的 Provider Simulator 跑通真实 PostgreSQL 上的纵向闭环。

当前实现聚焦 tokenized card 的 automatic capture，并把超时视为 `UNKNOWN`：系统会沿用同一个 Provider Operation 查询或等待 webhook，不会因为没有收到响应就盲目切换 PSP。Router v3 负责受约束、可解释的决策，Payment 保存资金状态与恢复预算，Provider Gateway 保存不可变连接身份和追加式外部证据。Simulator 不处理真实资金；接入 live PSP 仍需要真实 connector、密钥管理、安全合规与运营准备，因此这是一套可运行、可验证的支付核心，而不是可直接处理真钱的成品平台。

本版全书统一固定到配套实现 [bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)，对应 [CI run 33041147119（success）](https://github.com/YouEclipse/payments/actions/runs/33041147119)。每章与附录仍以 delivery manifest 的 claim→code→test 映射为准，不能仅凭全书级提交链接推导某项能力已经实现。

## 这本书适合谁

本书主要面向：

- 具有 3～8 年经验的后端工程师；
- 支付系统与 FinTech 工程师；
- 系统架构师和基础设施工程师；
- 希望进入支付领域的工程师；
- 具有工程背景、希望理解系统约束的支付产品经理。

默认读者已经了解 HTTP / REST API、数据库、Redis、消息队列、微服务，以及分布式系统中的基本一致性与可用性概念。书中会解释支付领域术语和资金语义，但不会从零讲解通用后端技术。

## 你将学到什么

读完本书后，读者应该能够：

- 解释 Customer、Merchant、PSP、Acquirer、Card Network 与 Issuer 如何共同完成一笔支付；
- 划分 Payment API、Payment Orchestrator、Payment Routing System 与 Provider Adapter 的职责边界；
- 为一次支付构建 Routing Context，并先通过 Eligibility 排除不可用路径；
- 设计可版本化、可验证、可灰度和可回滚的 Routing Rule Engine；
- 在样本量、数据延迟和选择偏差存在时评估 Provider 成功率；
- 综合成功率、成本、延迟、风险与健康状态做出可解释的 Routing Decision；
- 正确处理 timeout、unknown state、retry、fallback 与幂等，避免重复扣款；
- 建立 Routing Decision Store、核心指标、实验与审计能力；
- 让内部支付状态通过 webhook、状态查询和 Reconciliation 与 Provider 状态收敛；
- 设计支持多国家、多币种、多支付方式、多 PSP 和多 Region 的支付路由平台；
- 判断规则、统计方法、bandit 或 ML 各自适用的边界，而不是为了智能化而引入模型。

## 全书结构

### Part I — Payment Foundations

从一笔 Card Payment 的完整生命周期出发，建立 PSP、Acquirer、Card Network、Issuer、Authorization、Capture、Refund 与 Chargeback 等基本概念；随后从单 PSP 架构推导出 Payment Routing，并明确 Router 与 Orchestrator 的职责边界。

### Part II — Building the Routing Engine

逐步构建 Routing Context、规则模型、Provider Eligibility、成本模型、成功率估计、风险约束和 Provider Scoring，并通过统一的 Provider Adapter 隔离不同 PSP 的 API 与状态语义。

### Part III — Reliability

围绕支付最危险的失败场景展开：请求超时但资金操作可能已经成功。这里会建立 Payment、Attempt 与 Provider Request 的幂等关系，区分安全和不安全的 fallback，并设计 Provider Health 与 Circuit Breaker。

### Part IV — Operating Payment Routing at Scale

把路由从“能做决定”推进到“能够长期运营”：记录每次决定的证据，建立可观测性和 SLO，运行受控实验，通过 Reconciliation 修复状态差异，并将决定连接到真实费用、结算和拒付结果。

### Part V — Global Payment Routing

讨论多 Region、数据驻留、配置发布、安全与合规边界，以及什么时候可以在规则之上引入统计路由、bandit 或机器学习。

### Part VI — Case Study

以一个面向 10,000 个商户、50 个国家、30 种币种、20 种支付方式、15 家 PSP，并承受 100k TPS 峰值的 Global Payment Routing Platform 为综合案例，串联 API、状态机、存储、缓存、消息、决策、failover、reconciliation 和部署设计。

附录提供 Payment Intent / Attempt / Provider Request 状态模型与错误分类、安全合规和数据边界，以及容量规划与上线检查清单；[术语表](glossary.md)统一全书中英文支付术语。

## 从哪里开始

- 第一次接触支付系统：建议按[目录](SUMMARY.md)顺序阅读。
- 已有支付接入经验：可以先阅读样章[《为什么需要 Payment Routing》](part-1-foundations/why-payment-routing.md)，再进入 Routing Engine 与 Reliability 两部分。
- 正在设计多 PSP 平台：建议重点阅读系统边界、Eligibility、幂等、fallback、决策审计和 Reconciliation 相关章节。

## 项目维护

本书的后续修订统一遵循[写作与实现工作流](WORKFLOW.md)，把章节正文、Go 微服务增量、支付语义审计、图表配图、真实数据库验证和双仓库发布作为同一个交付闭环。每章与附录的固定代码、CI、claim→code→test 映射和能力边界可从[交付证据索引](deliveries/README.md)进入。
