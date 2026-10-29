# Designing Modern Payment Systems

## 从一笔支付到全球支付路由平台

现代支付系统面对的核心问题，不只是“如何调用一家支付服务商（PSP）的接口”，而是如何在多个国家、币种、支付方式、商户实体和服务商之间，持续做出安全、可解释、可运营的支付决策。

本书以**支付路由系统**为主线，从最简单的单一服务商架构出发。当业务逐步遇到覆盖差异、授权成功率、处理成本、风险约束、服务商故障和全球部署等问题时，再推导出路由上下文、规则引擎、服务商准入筛选、动态评分、后备路由、决策审计与对账等生产级能力。

这不是一本支付术语百科，也不会先给出一张复杂架构图再逐项解释组件。主要章节采用下面的演进方式：

```
问题 → 简单设计 → 规模化问题 → 生产级设计 → 权衡
```

读者会先看到一个足够简单的系统如何工作，再观察它在哪些真实条件下失效，最后理解新组件为什么出现、解决了什么问题，又带来了哪些代价。

> 中文版第一版包含 22 章、3 篇附录与术语表。阅读时请区分配套系统已经实现的能力、用于解释设计的 Go 教学模型，以及有待实现的目标方案；各章会说明适用范围与限制。英文版将在此基础上整体翻译并重写。

## 语言约定

正文以自然中文为主。支付概念优先使用“支付意图、支付尝试、资金操作、候选通道、准入筛选、路由决策、后备路由”等中文名称；仅在首次定义、术语表或需要对应外部规范时附英文。Go、PostgreSQL、HTTP、JSON、PSP、PCI DSS、3DS、品牌名、代码标识、接口字段和状态枚举保持原样，避免翻译损害工程精度。

## 代码示例约定

本书所有用于展示后端实现的示例代码统一使用 **Go**，包括领域模型、接口、支付状态机、路由规则求值、服务商适配器、幂等与重试、数据访问以及并发处理。示例会优先使用清晰的类型和标准库，并尽可能保持可编译、可测试，而不是只给出无法运行的片段。

JSON、SQL、Mermaid 和少量伪代码仍会分别用于表达接口数据、存储模型、架构时序和决策公式；它们只承担对应的描述职责，不作为另一套实现语言。

## 配套 Go 支付系统

本书同步维护一个可运行的微服务实现：[YouEclipse/payments](https://github.com/YouEclipse/payments)。它不是把章节里的片段简单拼在一起，而是用支付服务、路由服务、服务商网关和仅限开发与测试的服务商模拟器跑通真实 PostgreSQL 上的纵向闭环。

当前实现聚焦令牌化银行卡的自动请款，并把超时视为 `UNKNOWN`：系统会沿用同一个服务商资金操作进行查询或等待回调通知，不会因为没有收到响应就盲目切换 PSP。路由服务 v3 负责受约束、可解释的决策，支付服务保存资金状态与恢复预算，服务商网关保存不可变连接身份和追加式外部证据。模拟器不处理真实资金；接入真实生产 PSP 仍需要服务商连接器、密钥管理、安全合规与运营准备，因此这是一套可运行、可验证的支付核心，而不是可直接处理真钱的成品平台。

本版全书统一固定到配套实现 [bc4afde24da708d22f67da72a90a1557b1477173](https://github.com/YouEclipse/payments/commit/bc4afde24da708d22f67da72a90a1557b1477173)，对应[持续集成任务 33041147119（成功）](https://github.com/YouEclipse/payments/actions/runs/33041147119)。各章与附录会分别说明当前实现、教学模型与未来方案的边界；通过测试的配套系统并不代表书中讨论的所有能力都已实现。

## 这本书适合谁

本书主要面向：

* 具有 3～8 年经验的后端工程师；
* 支付系统与金融科技工程师；
* 系统架构师和基础设施工程师；
* 希望进入支付领域的工程师；
* 具有工程背景、希望理解系统约束的支付产品经理。

默认读者已经了解 HTTP、REST API、数据库、Redis、消息队列、微服务，以及分布式系统中的基本一致性与可用性概念。书中会解释支付领域术语和资金语义，但不会从零讲解通用后端技术。

## 你将学到什么

读完本书后，读者应该能够：

* 解释顾客、商户、PSP、收单机构、卡组织与发卡行如何共同完成一笔支付；
* 划分支付接口、支付编排器、支付路由系统与服务商适配器的职责边界；
* 为一次支付构建路由上下文，并先通过准入筛选排除不可用路径；
* 设计可版本化、可验证、可灰度和可回滚的路由规则引擎；
* 在样本量、数据延迟和选择偏差存在时评估服务商成功率；
* 综合成功率、成本、延迟、风险与健康状态做出可解释的路由决策；
* 正确处理超时、未知状态、重试、后备路由与幂等，避免重复扣款；
* 建立路由决策存储、核心指标、实验与审计能力；
* 让内部支付状态通过回调通知、状态查询和对账与服务商状态收敛；
* 设计支持多国家、多币种、多支付方式、多 PSP 和多区域的支付路由平台；
* 判断规则、统计方法、上下文多臂老虎机或机器学习各自适用的边界，而不是为了智能化而引入模型。

## 全书结构

### 第一部分：支付基础

从一笔银行卡支付的完整生命周期出发，建立 PSP、收单机构、卡组织、发卡行、授权、请款、退款与拒付等基本概念；随后从单一服务商架构推导出支付路由，并明确路由服务与支付编排器的职责边界。

### 第二部分：构建路由引擎

逐步构建路由上下文、规则模型、服务商准入筛选、成本模型、成功率估计、风险约束和服务商评分，并通过统一的服务商适配器隔离不同 PSP 的接口与状态语义。

### 第三部分：可靠性

围绕支付最危险的失败场景展开：请求超时但资金操作可能已经成功。这里会建立支付、支付尝试与服务商请求的幂等关系，区分安全和不安全的后备路由，并设计服务商健康度与熔断器。

### 第四部分：规模化运营支付路由

把路由从“能做决定”推进到“能够长期运营”：记录每次决定的证据，建立可观测性和 SLO，运行受控实验，通过对账修复状态差异，并将决定连接到真实费用、结算和拒付结果。

### 第五部分：全球支付路由

讨论多区域、数据驻留、配置发布、安全与合规边界，以及什么时候可以在规则之上引入统计路由、上下文多臂老虎机或机器学习。

### 第六部分：综合案例

以一个面向 1 万个商户、50 个国家、30 种币种、20 种支付方式、15 家 PSP，并承受 10 万 TPS 峰值的全球支付路由平台为综合案例，串联接口、状态机、存储、缓存、消息、决策、故障转移、对账和部署设计。

附录提供支付意图、支付尝试与服务商请求的状态模型和错误分类、安全合规与数据边界，以及容量规划与上线检查清单；[术语表](glossary.md)统一全书中英文支付术语。

## 从哪里开始

* 第一次接触支付系统：建议按[目录](SUMMARY.md)顺序阅读。
* 已有支付接入经验：可以先阅读样章[《为什么需要支付路由》](part-1-foundations/why-payment-routing.md)，再进入路由引擎与可靠性两部分。
* 正在设计多 PSP 平台：建议重点阅读系统边界、准入筛选、幂等、后备路由、决策审计和对账相关章节。

##

### Working Subtitle

Payment Routing, Reliability, Cost Optimization, and Risk Control at Scale

### Book Positioning

Designing Modern Payment Systems is a practical system design book about building modern payment infrastructure, with payment routing as the central system. The book explains how payment platforms route transactions across PSPs, acquirers, payment methods, regions, currencies, cost models, risk signals, and reliability constraints.

### Target Readers

The primary readers are backend engineers, payment engineers, fintech architects, infrastructure engineers, and technical product managers who want to understand how modern payment routing systems are designed.

The reader is expected to understand basic backend systems, APIs, databases, queues, and distributed systems concepts. Prior payment experience is helpful but not required.

### Core Promise

After reading this book, the reader should be able to design a payment routing system that is reliable, observable, cost-aware, risk-aware, and scalable across multiple payment providers and markets.

### Core System

The central system in this book is the Payment Routing System. It receives a payment request and decides where and how the transaction should be processed.

A routing decision may consider provider availability, payment method, card brand, issuer country, merchant configuration, transaction amount, currency, cost, historical success rate, risk signals, compliance constraints, latency, and fallback options.

### Scope

This book covers payment routing, PSP integration, payment lifecycle, rule-based routing, cost-aware routing, risk-based routing, retries, fallbacks, observability, reconciliation, settlement implications, experimentation, and global payment infrastructure.

### Non-Goals

This book is not a general introduction to consumer payments, a legal compliance manual, a PCI implementation checklist, or a complete accounting textbook. Topics such as PCI, ledger design, settlement, and reconciliation are covered only to the extent that they influence payment routing system design.

### Writing Principles

Each chapter should connect domain concepts with system design decisions.

Each major concept should answer four questions:

1. What problem does this solve?
2. What can go wrong?
3. What does a simple design look like?
4. How does the design evolve at scale?

The book should prefer concrete architecture, data models, state machines, failure modes, and operational trade-offs over abstract explanations.
