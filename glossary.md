# 术语表

本术语表服务于全书的 Payment Routing 主线。定义强调术语在架构和路由决策中的含义；不同国家、卡组织和 Provider 的具体规则可能不同，实际实现应以合同、接口文档及适用法规为准。

## 数字与符号

### 三维安全认证（3-D Secure，3DS）

银行卡线上支付的持卡人认证协议。Provider 对 3DS 版本、Frictionless / Challenge 流程，以及 SCA exemption 信息传递的支持能力，可能成为 Eligibility 或 Risk-Aware Routing 的输入；Frictionless 仍是认证流程，不等同于免除 SCA。

### 强客户认证（Strong Customer Authentication，SCA）

要求基于多个独立要素验证付款人的认证机制，在欧洲支付场景中尤其常见。Router 通常消费上游给出的认证要求，并选择能够执行相应流程的通道，而不自行判定法律适用性。

## A

### 收单机构（Acquirer）

为 Merchant 提供银行卡收单能力，并将交易送入 Card Network 的持牌机构。PSP 可能直接是 Acquirer，也可能连接一个或多个 Acquirer。

### 主动-主动（Active-Active）

多个 Region 同时承载生产流量的部署方式。用于支付路由时，必须额外处理 Intent 所有权、全局幂等、配置版本和跨区 Fallback，不能只复制无状态服务。

### 实际成本（Actual Cost）

来自 Provider 结算行、费用明细或已确认经济事实的成本。必须保存来源、币种、匹配范围和修订关系，不应覆盖决策时的 Expected Cost；尚未到账或匹配不完整的值不能伪装成最终成本。

### 授权（Authorization）

对一笔支付请求进行批准或拒绝的过程，通常由 Issuer 决策；在特定条件下，Card Network 或处理方也可能按既定规则进行 stand-in processing。批准通常表示额度或资金被预留，并不等于 Merchant 已经收到资金，也不等于后续 Settlement 必然完成。

### 授权成功率（Authorization Rate）

Authorization 被批准的比例。分子和分母是否包含技术失败、重复尝试、风控拦截等必须明确；用于 Routing 时还需指定 Provider、国家、Issuer、Payment Method、时间窗口和样本量。

### 地址验证服务（Address Verification Service，AVS）

将付款人提供的地址信息与 Issuer 侧记录进行匹配的校验能力。AVS 的覆盖和结果语义因市场及 Provider 而异，可作为风控输入或通道能力，不能被视为全球一致的布尔字段。

## B

### 银行识别号 / 发卡机构识别号（BIN / IIN）

卡号前缀中用于识别发卡机构、卡产品等信息的部分。BIN/IIN 可帮助细分 Provider Performance，但属于需要谨慎处理的支付数据；长度和元数据不应被硬编码为固定假设。

### 舱壁隔离（Bulkhead）

把连接池、线程、并发额度或队列按 Provider 等故障域隔离，避免一家 Provider 的延迟耗尽整个 Router 的资源。

## C

### 持卡人数据（Cardholder Data，CHD）

PCI DSS 定义的受保护账户数据，至少包含完整 PAN；与 PAN 一同出现时，还可能包括持卡人姓名、有效期和服务代码。CHD 可以在有业务必要并满足适用控制时保存，不应与 Authorization 后禁止保存的 SAD 混为一谈。

### 卡验证值（CVV / CVC / CID）

用于 Card-not-present 等场景验证持卡人掌握卡片的安全码。它属于 SAD，在 Authorization 完成后不得保存，即使经过加密也不例外；Router 不应把它写入 Context、Decision、日志、Trace 或消息。

### 候选通道（Route Candidate）

可能执行当前 Payment Attempt 的具体通道，通常由 Provider、Merchant Account、Connector、Region 等共同标识。Scoring 比较的是 Candidate，而不只是抽象的 PSP 名称。

### 能力目录（Capability Catalog）

记录每条 Candidate 支持的国家、币种、Payment Method、Capture Mode、认证方式等能力的版本化数据。它是 Eligibility 的重要输入，但不等同于动态 Provider Health。

### 连接修订（Connection Revision）

对一份执行连接配置的不可变引用。Attempt、Operation 和 Gateway 固定同一 ID/revision，配置轮换创建新修订，不能把旧修订重新绑定到另一账户或 endpoint；旧 UNKNOWN 不随新配置迁移。

### 配置包（Config Bundle）

将 Rule、Capability、Fee、Experiment、Model 等不可变 artifact 引用和摘要固定在一起的发布单元。验证、审批与运行实例加载必须针对同一个 bundle；运行时原子切换，不能逐个替换半套配置。

### 请款（Capture）

在 Authorization 之后确认全部或部分已授权金额、进入 Clearing/Settlement 的操作。自动 Capture 与手动 Capture 的支持差异会影响 Provider Eligibility。

### 卡组织（Card Network / Card Scheme）

连接 Acquirer 与 Issuer、定义卡交易规则并传递授权和清算信息的网络，例如 Visa、Mastercard。其费用和规则可能通过 Provider 合同间接影响路由成本。

### 拒付（Chargeback）

持卡人争议等原因触发、通过卡组织流程向 Merchant 追索资金的机制。Chargeback 可能在支付数周或数月后发生，因此 Risk-Adjusted Cost 存在延迟反馈。

### 熔断器（Circuit Breaker）

在 Provider 故障达到阈值后暂时阻止新请求，并在恢复阶段以有限探测判断是否重新放量的机制。熔断状态应按真实故障维度建模，而非总是全局开关。

### 清算（Clearing）

参与方交换、校验和汇总交易明细并计算应收应付的过程，通常发生在 Authorization 之后、Settlement 之前。

### 连接器（Connector）

某套 Provider 协议、API 版本和字段/状态映射的实现。它不是包含账户、区域、endpoint 与 credential scope 的完整 Provider Connection；同一 Connector 实现可以服务于多个精确连接，但不能自行替换连接或路由。

### 控制平面（Control Plane）

管理 Rule、Cost、Capability、Model 和 Experiment 的创建、验证、审批及发布的系统。它可与实时承载 Routing Decision 的 Data Plane 分离。

### 跨境费（Cross-Border Fee）

交易参与方处于不同国家或地区时可能产生的额外费用。其判定依据不一定只是 Customer Country，必须按 Provider 与卡组织的实际规则建模。

## D

### 数据平面（Data Plane）

在支付在线链路中构建 Context、过滤 Candidate、执行 Rule/Scoring 并返回 Routing Decision 的组件集合，通常有严格的延迟与可用性预算。

### 数据驻留（Data Residency）

要求特定类别的数据在指定司法辖区内存储或处理的约束。它会影响 Region 选择、Context 字段、Decision Log 和离线指标汇总方式。

### 拒绝（Decline）

Authorization 未获批准的业务结果。Issuer / Network Decline 应与授权前的 Provider Risk Block、Gateway Rejection 和 Integration Error 分开归因。Decline Code 需要经 Adapter 归一化，但应保留 Provider 原始码；“软拒绝”和“硬拒绝”的可重试性必须结合 Payment Method、Card Network advice 及 Provider 语义判断。

### 决策原因码（Decision Reason Code）

结构化表达 Candidate 被排除、加减分或最终胜出的原因。它比自由文本日志更适合审计、指标聚合和稳定的 API 契约。

### 摘要（Digest）

对明确规范化后的 Context、artifact 或 Decision 计算的内容指纹。摘要能检查内容一致性，但本身不证明发布者身份、审批授权或事实正确性；不能用相同版本名掩盖不同内容。

## E

### 适格性过滤（Provider Eligibility）

根据能力、合同、合规和不可用状态等硬约束，从全部 Provider 通道中筛出可以处理当前交易的 Candidate Set。`Unavailable` 或 `CircuitOpen` 可以构成硬排除；`Degraded` 通常作为评分惩罚或流量限制。Eligibility 先于偏好评分执行。

### 有效样本量（Effective Sample Size，ESS）

衡量加权样本实际信息量的诊断量，常用形式为 `(Σw)² / Σw²`。少数极大权重可让记录很多但 ESS 很小；ESS 不能修复缺少策略支持、错误 propensity 或未成熟结果。

### 实验暴露（Experiment Exposure）

业务单元真正进入实验处理的事件。Assignment 仅表示分桶，不等于候选已执行；shadow 也不产生资金暴露。指标需分别保存分配、决策、执行和结果，避免以“只统计成功送达的请求”选择样本。

### 期望成本（Expected Cost）

在支付结果最终确定前，根据费率、成功概率、FX、失败尝试和潜在风险等估算的成本。它应与结算后得到的 Actual Cost 区分并持续校准。

## F

### 故障转移路由（Fallback Routing）

前一次 Attempt 得到明确且允许重新路由的结果后，选择另一条 eligible route 创建新 Attempt 的恢复过程。同一 Provider 上使用相同 operation identity 的重试属于 Retry，不是 Fallback；Unknown 状态下不能直接跨 Provider Fallback。

### 隔离令牌（Fencing Token / Ownership Epoch）

随执行所有权转移而递增的代数。写事务和副作用入口验证当前 epoch，拒绝旧 owner；DNS、租约过期或 worker 存活都不能替代该检查。它不能撤回已经发到 Provider 的请求，仍须恢复原 Operation。

### 外汇转换（Foreign Exchange，FX）

不同币种之间的换算过程及其汇率和加价。FX 发生在哪一方、以何种汇率结算，会影响 Routing 的真实成本和 Merchant 到账金额。

### 欺诈引擎（Fraud Engine）

评估交易欺诈风险并输出 Decision、Score 或 Reason 的系统。Payment Router 可以消费其结果，但通常不应复制模型特征或接管欺诈决策职责。

## G

### 支付网关（Payment Gateway）

接收 Merchant 支付请求、完成安全接入或协议转换并连接后端处理方的系统。Gateway、PSP、Orchestrator 和 Router 的功能可能在商业产品中重叠，本书按架构职责而非品牌名称区分它们。

### 护栏（Guardrail）

在优化或实验之前不可违反的约束，例如合规 Eligibility、风险上限、Provider 容量和最大 Fallback 次数。Guardrail 不应被更高的模型分数抵消。

## H

### 硬拒绝（Hard Decline）

通常表示不应原样重试的业务拒绝，例如被报告为无效或禁止的支付凭证。但分类并非跨 Provider 完全统一，路由系统应根据归一化码和明确 Policy 决定动作。

### Provider 健康状态（Provider Health）

由延迟、超时、技术错误和主动探测等信号形成的运行时状态。它应带维度、时间戳和置信度，不能与 Authorization Rate 简单等同。

### 归属区域（Home Region）

负责某个 Intent 权威状态与执行归属的区域。它不同于接收 HTTP 的 ingress Region，也不同于实际 worker 所在 execution Region 和外部 Provider Region。流量转移不自动转移资金写入权。

## I

### 幂等（Idempotency）

同一业务操作被重复提交时，不产生额外非预期副作用的性质。Provider 可能返回首次响应、对象的当前状态或冲突错误，因此响应语义不一定逐字一致；可靠实现需要业务身份、持久化状态和并发约束，而不只是缓存响应。

### 幂等键（Idempotency Key）

调用方为一次逻辑操作提供的稳定标识。系统必须定义它的作用域、有效期、参数是否可变，以及相同 Key 对应不同请求时的错误语义。

### 完整性状态（Integrity Status）

表示聚合内证据是否能够安全自动解释的独立维度。本书使用 `healthy` / `review_required`；进入 review 不抹掉已 captured 事实，也不是把 `succeeded` 改成 `failed`。

### Inbox / Outbox

分别用于记录已接收消息身份与事务内待发布事件的持久化模式。与领域状态在同一事务中提交，处理进程崩溃窗口；投递仍可能重复，下游仍须按事件 ID 和内容验证幂等，不能宣称端到端恰好一次。

### 发卡机构（Issuer）

向 Customer 发行银行卡或支付账户，并在 Authorization 时根据账户状态、资金、风险和规则批准或拒绝交易的机构。

## L

### 账本（Ledger）

以可审计分录记录资金权利义务变化的系统。Router 可以引用支付与结算结果，但不应把 Routing Decision Store 当作资金 Ledger。

### 最近已验证配置（Last-Known-Good，LKG）

最近一次通过完整校验并成功加载的配置。控制平面暂时不可用时可按有效期与安全边界继续使用；配置过期、强制停用或缺少必需 artifact 时，不能以 LKG 名义无限运行。

## M

### 商户（Merchant）

向 Customer 提供商品或服务并发起收款的主体。Merchant 的合同、业务类别、国家和策略范围都会影响 Candidate Eligibility 与 Routing Policy。

### Provider 商户账户（Merchant Account）

Merchant 或支付平台在某个 Provider 下用于处理交易的账户或合同实体。费率、币种能力、限额和风险策略通常绑定到它，而不是只绑定 Provider 名称。

### 支付方式（Payment Method）

Customer 用于完成支付的工具或网络，例如银行卡、数字钱包、银行转账和本地支付方式。不同方式的同步性、可重试性、认证和最终性差异会直接约束路由。

## O

### 离线策略评估（Off-Policy Evaluation，OPE）

使用历史行为策略的数据估计另一目标策略的结果。需明确 reward、成熟窗口、记录的行动概率和支持集合；目标策略会选择但历史概率为零的行动，不能靠重加权识别。结果不是未经线上受控实验验证的因果提升承诺。

### 支付编排器（Payment Orchestrator）

管理 Payment Intent 生命周期并协调 Risk、Router、Provider Adapter、异步事件和恢复流程的组件。Router 负责产生不可变 Route Plan 与执行 guardrail-only revalidation；Payment Orchestrator 决定是否基于旧 Attempt 证据创建新资金路径，不把 binding 冒充新的 Router 排名。

## P

### 主账号（Primary Account Number，PAN）

用于识别银行卡账户的号码。PAN 属于 CHD，不等同于 SAD；本书 Routing Context 只接收必要派生属性，不携带 PAN、CVV 或可用于执行支付的 Token。Token 只在受控 Payment/Gateway 边界流动，不能因其不是原始卡号就进入日志或 Decision。[PCI SSC Glossary](https://www.pcisecuritystandards.org/glossary/)

### 支付尝试（Payment Attempt）

为完成同一个 Payment Intent 而进行的一次受控执行，固定支付方式与 Provider Connection。对同一逻辑 Operation 的网络 Retry 不创建新 Attempt；只有前一 Attempt 已明确无副作用且策略允许的 Fallback 才能创建新 Attempt。UNKNOWN 阻断新资金路径。

### 支付意图（Payment Intent）

Merchant 希望收取某个金额与币种的一次业务目标。它可以经历认证、多个 Attempt、Capture 或失败恢复，生命周期通常由 Orchestrator 持有。

### 支付服务提供商（Payment Service Provider，PSP）

向 Merchant 或平台提供一种或多种支付接入、处理、收单、Payment Method 聚合等服务的机构。PSP 是宽泛商业术语，其具体能力应通过 Capability 和 Connector 建模。

### 支付处理方（Payment Processor）

为交易提供消息处理、连接或后台处理能力的机构或系统。它可能与 PSP 或 Acquirer 是同一实体，也可能只是链路中的技术服务方。

### 支付凭证 Token 化（Tokenization）

用受控 Token 代替 PAN 等敏感凭证的技术。Token 的可用范围可能受 Provider、Merchant、Network 或 Region 限制，这种可移植性会影响 Fallback。

### Provider Adapter

把平台统一的支付操作映射为某个 Provider API，并归一化响应、错误类别和 webhook 事件的适配层。Adapter 必须显式暴露 Idempotency 的作用域、保留期、并发和响应语义差异，不能用统一接口将其隐藏；它执行已选路线，不应暗中实施独立 Routing Policy。

### Provider Connection

可精确执行的一份 Provider / Merchant Account / Connector / Region 配置，包含由 Gateway 掌握的 endpoint 与 credential scope。Router 返回非秘密引用，Payment 固定 ID、revision 与外部 namespace，执行层不能临时二次选择。

### Provider Namespace

Adapter 根据供应商契约声明的外部身份冲突域，不能只按 Provider 品牌名推断。Reference、event、幂等键若有不同作用域，应分别表达；同 namespace 的连接不因此共享 webhook 认证权限。

### Provider Operation

一次逻辑资金操作，如 SALE、AUTHORIZE、CAPTURE、VOID、REFUND。内部 Operation ID、Provider key、语义 payload 和精确执行连接固定；网络请求可多次，资金身份仍是一个。正文泛称 Provider Request 时需区分逻辑操作与一次传输。

### 行动概率（Propensity）与支持集合（Support）

Propensity 是行为策略在当时 Context 下选择某行动的概率，不是该行动的预测成功率。OPE 要求目标策略的整个行动支持集合都被日志策略覆盖，而不只是检查已观察行动的概率大于零。

## R

### 对账（Reconciliation）

将平台内部 Payment/Provider State 与 Provider API、webhook、交易或结算文件进行匹配，发现 Missing、Duplicate、Status/Amount Mismatch 并触发受控修复的过程。

### 退款（Refund）

Merchant 对已 Capture 的全部或部分金额发起返还的操作。Refund 与 Void 的时间点和网络语义不同，也可能产生单独费用和异步状态。

### 重试（Retry）

对同一逻辑操作再次发起执行。支付 Retry 固定原 Operation、Provider/账户/Region/修订、幂等键与语义 payload，并受结果确定性、Provider 契约、总 Deadline 与 Retry Budget 约束；它不创建新 Attempt，也不等同于 Fallback。

### 风险感知路由（Risk-Aware Routing）

把 Risk Decision、认证要求和 Provider 风控能力纳入 Eligibility 或 Scoring 的路由方式。其目标是在风险系统允许的范围内选择执行路径，而不是由 Router 重新做欺诈判断。

### 路由上下文（Routing Context）

由 Payment ContextBuilder 为一次决策固定的版本化最小输入，包含业务字段和必要非秘密派生属性。每个观察值带来源、版本、时间与 known/unknown/stale 等状态；缺失不能伪装成默认真值。动态 Health/Performance 可用独立只读快照参与决策，但都需记录版本，不能混入未审计的实时查询。

### 路由决策（Routing Decision）

Router 给出的 Candidate 选择或明确负向 outcome，以及对应理由与版本证据。`selected` 分支可以形成不可变 `RoutePlanDecision`；`no_eligible_candidate`、`blocked`、`unsupported`、`requires_review` 或依赖不可用等负向分支不得伪造 Attempt、Selection 或 Provider Operation。

### 路由计划决策（RoutePlanDecision）

Router-authored 的不可变执行许可边界，固定 candidate order、Policy/Capability/Feature 版本、最大 Attempt、预算与有效期。Fallback 的 guardrail-only revalidation 只能 veto 固定下一项，不能新增候选、重排、重评分或延长计划；过期或证据不可用时 Stop。

### Fallback 执行绑定（FallbackExecutionBinding）

Payment Orchestrator 在获准 successor transaction 中写入的不可变对象，引用原 `RoutePlanDecision`、前序 Attempt、固定 next Candidate、新鲜 Risk/guardrail 证据与预算变化。它证明 Payment 为什么创建该 Attempt，但不是 Router Decision，也不授予 Plan 外执行。

### 路由决策记录（Routing Decision Record）

用于解释和复现 Routing Decision 的不可变事实记录，包含候选、排除原因、分数分解、版本引用和时延等，但不应复制无必要的敏感数据。

### 路由策略（Routing Policy）

决定如何过滤、偏好、分流或选择 Candidate 的业务规则与优化目标。Policy 的执行属于 Data Plane，其创建、审批和发布通常属于 Control Plane。

### 路由规则引擎（Routing Rule Engine）

按确定性语义执行版本化 Routing Rule 的组件，负责条件匹配、优先级和冲突处理。Rule Engine 不是完整 Router，还需与 Eligibility、Feature、Scoring 和 Selection 配合。

### 路由配置平台（Routing Configuration Platform）

为 Routing Policy 及相关配置提供编辑、验证、回放、审批、发布、灰度和回滚能力的控制平面系统。

### 恢复点目标 / 恢复时间目标（RPO / RTO）

分别表达可接受的数据丢失窗口与恢复服务的目标耗时。支付 failover 必须单独评估执行身份、幂等与资金证据的 RPO；总体数据库复制延迟不是可安全重放未知资金操作的证明。

## S

### 敏感认证数据（Sensitive Authentication Data，SAD）

用于验证持卡人的安全数据，包括完整磁道数据、CVV/CVC/CID 以及 PIN/PIN Block。按照 PCI DSS，这些数据在 Authorization 完成后不得保存，即使经过加密也不例外；它们不应进入 Routing Context、Decision Store、日志或消息队列。

### 结算（Settlement）

参与机构根据 Clearing 结果完成净额或总额资金转移的过程。它通常晚于 Authorization 和 Capture，到账时间与费用受 Provider、币种和市场规则影响。

### 卡组织费（Scheme Fee）

Card Network 对交易或相关服务收取的费用。其结构可能复杂且通过 Acquirer/PSP 转嫁，Cost Model 需要以实际合同和结算数据校准。

### 评分（Provider Scoring）

在已满足硬约束的 Candidate 之间，根据成功率、成本、延迟、风险等目标计算相对偏好的过程。评分不能覆盖 Eligibility 和 Guardrail。

### 影子决策（Shadow Decision）

在真实请求的相同快照上计算 challenger 决策，用于比较分歧和可运行性，但不 dispatch、不创建第二 Operation，也不把未执行候选的结果当作已观察结果。

### 副作用确定性（Side-Effect Certainty）

描述操作是否确定尚未发生、可能发生或已确认发生的证据维度。它与 HTTP 错误、业务成功率和 Retryable 不同；`retryable=true` 不代表可以换 PSP，`not_found` 也只有在明确查询契约下才可能成为无副作用证据。

### 结算行（Settlement Line）

来自 Provider 报表的不可变经济事实，携带文件/行身份、外部 namespace、reference、金额、币种和类型。重复同内容幂等，更正追加新事实；它不是 Payment Intent 的新状态，也不是会计分录的替代品。

### 软拒绝（Soft Decline）

可能在补正错误数据、完成所需认证或根据 Network / Issuer advice 稍后再试后成功的业务拒绝。是否允许 Retry/Fallback 必须由标准化 Decline Reason 和明确 Policy 决定，而非只看 HTTP 状态码；不能为了获得批准而任意改变原交易语义。

### 静态路由（Static Routing）

按固定 Provider、固定优先级或简单比例选择通道的方式。它易于实现和解释，适合作为系统起点或安全降级策略，但无法适应细粒度表现变化。

### 基于成功率的路由（Success-Rate Routing）

根据与当前交易相关的历史 Performance 估计各 Candidate 成功概率并参与选择的方式。可靠实现必须处理分桶、样本量、噪声、Cold Start 和数据延迟。

## T

### 流量分配（Traffic Splitting）

按确定性规则把部分交易分配给不同 Provider 或策略的机制，可用于合同配额、容量控制或实验。实验分流应保持同一业务单位的分组稳定。

### 交易（Transaction）

泛指一笔支付相关业务或网络记录。由于不同系统可能用它指 Payment Intent、Attempt、Authorization 或 Settlement Line，数据模型和接口中应使用更具体的对象名称。

## U

### 未知状态（Unknown State / Unknown Result）

调用方无法确认 Provider 是否已产生副作用的状态，例如 Provider 已批准但响应丢失。Unknown 不是普通失败；在 Query、webhook 或 Reconciliation 确认前贸然 Fallback 可能造成重复支付。

## V

### 撤销（Void）

在原交易完成 Settlement 前、且 Provider 状态仍允许时取消交易的操作。部分 Provider 用 Void 专指释放未 Capture 的 Authorization，另一些也允许取消已 Capture 或已提交但尚未结算的交易；Adapter 需要区分 authorization reversal、cancel 与 void，并归一化各自的状态窗口和结果语义。

## W

### Webhook

Provider 主动向平台发送异步状态事件的 HTTP 回调。接收方必须验证来源，并处理重复、乱序、延迟和投递失败；Webhook 不能被假设为恰好一次且严格有序。

## 与路由策略相关的复合术语

### 成本感知路由（Cost-Aware Routing）

把固定费率、比例费率、跨境费、FX 及失败尝试等 Expected Cost 纳入选择的路由方式。最低单次报价不一定带来最低 Cost per Successful Payment。

### 智能路由（Intelligent Routing）

使用统计估计、Contextual Bandit 或 ML 等方法动态优化路由的统称。智能路由仍必须受 Eligibility、合规、风险、容量和可靠性 Guardrail 约束。

### 基于规则的路由（Rule-Based Routing）

使用声明式 Condition 与 Action 表达 Merchant、国家、币种等确定性 Policy 的路由方式。生产系统还需要规则版本、冲突检测、验证、审批和回滚能力。
