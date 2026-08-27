# 《Designing Modern Payment Systems》全书交付合同

> 性质：写作与配套 Go 系统的执行地图，不是章节正文，也不代表下列能力已经实现。
> 覆盖范围：第 4～22 章与附录 A～C；第 1～3 章是前置基线。目录编号保持不变。

## 0. 合同状态

本文件描述“应该交付什么”，默认状态是 `target_contract`，不是完成声明。每项能力只能按以下三种状态解释：

| status | 可声称的范围 | 所需证据 |
| --- | --- | --- |
| `implemented` | 当前 `payments` 代码中存在并由实际测试覆盖的窄能力 | 对应 delivery manifest 的 claim→code→test 映射、固定实现 commit 与该 SHA 的绿色 CI run |
| `teaching_model` | 正文中可编译、可反例验证的 Go 模型 | 本书 verifier/专属 tests；不能推导服务、数据库、worker 或 live Provider runtime 已实现 |
| `target_contract` | 未来实现必须满足的 authority、状态与失败语义 | 设计说明和失败测试计划；没有 runtime 完成含义 |

章节或附录的当前 status 以各自 delivery manifest 为准；同一章可以同时包含已实现基线、教学模型和目标合同。下文“窄而完整的 Go 纵向增量”是目标描述，只有 manifest 明确标为 `implemented` 且固定代码证据齐全后，才能改称当前实现。

## 1. 使用方式与全书范围护栏

每章围绕一条“能从 API 进入、穿过领域模型与持久化、在失败条件下仍可验证”的窄纵向切片组织目标合同。章节正文遵循 `Problem → Simple Design → Problems at Scale → Production Design → Failure Modes → Trade-offs → Summary`；需要配套 runtime 的能力必须代码先于书稿发布，代码 CI 全绿后才把固定 SHA 写入对应 delivery manifest。

以下两项不是独立章节：

- **Provider Adapter 与 Provider connection 身份**并入第 3 章；第 6 章只扩展 Adapter 对外发布的版本化能力事实与 Eligibility，不另增“Adapter 章”。
- **Risk-Aware Routing**并入第 9 章 Provider Scoring；Fraud Engine 继续拥有风险判断，Router 只消费带版本、时效和原因码的 Risk Decision。

全书共同红线：

1. Authorization、Capture、Settlement 是不同资金阶段；`SALE` 只是当前 automatic-capture 的 Provider 原生操作，不能被写成通用 Authorization。
2. Capture、Void、Refund 始终绑定原 Provider connection 和 transaction reference，不重新路由。
3. timeout、断连、响应丢失和含糊 5xx 先进入 `UNKNOWN`；同一逻辑资金操作只能复用原 connection、payload、Provider key 与 operation ID。
4. Eligibility 硬约束先于 Rule 偏好和 Scoring；分数、实验或模型都不能复活不合格 Candidate。
5. Router 生成可解释 Decision；Payment/Orchestrator 拥有 Intent、Attempt、Operation、幂等和资金状态迁移；Gateway/Adapter 执行已固定的 connection。
6. 所有迟到资金证据都追加保存，不能 last-write-wins；重复或更弱且不冲突的观察按领域归约 no-op，只有不相容、身份冲突或无法安全归属的证据才进入 `review_required`。
7. Tokenization 不自动移出 PCI DSS scope；PAN、CVV/SAD、Provider secret 不进入 Routing Context、Decision、日志、Trace、Metric 或事件。

## 2. 当前代码基线与第 3 章前置合同

当前实现的审计锚点如下；后续章节必须在这些真实边界上演进，不另造一套示例系统。

| 当前事实 | 工程锚点 | 后续约束 |
| --- | --- | --- |
| Payment 是支付聚合与商户幂等的唯一写入者，并原子固定 v3 Context/Decision/Risk/Optimization 与不可续期执行期限 | `services/payment/internal/app/service.go`、`services/payment/internal/postgres/{routing_evidence,recovery}.go`、迁移 `005`/`006` | Intent/Attempt/Operation 的迁移仍在 Payment 事务内完成；首次执行权限和已 dispatch recovery 不得混为一谈 |
| Router v3 已按精确 connection/revision 运行 Capability → Rule → Risk → Cost/observational Performance → integer Scoring 的约束流水线 | `services/router/internal/{capability,decision,cost,performance,scoring}`、`contracts/internal/router-api.openapi.yaml` | fixture-backed 本地 evidence 不得冒充真实 PSP 报价、因果成功率、动态 Health 或 Fraud Engine |
| Gateway 以不可变 connection revision 执行 Provider Operation，使用 dispatch CAS、结构化 absence proof 与追加式 observation ledger 收敛不确定结果 | `services/provider-gateway/internal/gateway/`、迁移 `0003`～`0005`、`contracts/internal/provider-gateway-api.openapi.yaml` | Adapter 不能暗中改路；query/resubmit 必须使用原 operation、connection、key 与 payload，普通 404 不授权资金重提 |
| 当前资金切片仅支持 tokenized card、automatic capture、单个 `SALE` | `contracts/public/payment-api.openapi.yaml`、`README.md` | manual capture、void、refund 只有在相应章节明确实现后才能宣称支持 |
| PostgreSQL 已有 merchant idempotency、outbox/inbox、reference uniqueness、quarantine、不可重置 recovery budget/deadline/lease、路由快照与时钟见证 | `services/payment/internal/postgres/`、`docs/runbooks/` | Redis、broker 或新服务只有在章节压力真实要求时引入；不能替代数据库事实，生产升级必须执行 006 预检 |
| CI E2E 证明 approved、declined、Risk/no-candidate 零资金路径、lost response→UNKNOWN→同 operation 收敛，以及并发幂等只有一条资金路径 | `test/e2e/payment_flow_test.go`、`.github/workflows/ci.yml` | 后续测试必须保留这些回归；当前 Simulator 证据不等于真实 PSP certification |

第 3 章结束时必须提供后续章节可依赖的最小前置：

- `RouteCandidateID` 稳定标识 `Provider + Merchant Account + Connector + Execution Region`；Router response、Payment Attempt、Provider Operation、Gateway request 与数据库都保存该身份或不可变快照。
- Router 只返回 Decision；Payment 固定所选 connection 后创建 Attempt/Operation；Gateway 根据 connection 找到 Adapter 与凭证，不能仅按 Provider 名称临时选择账户或 Region。
- Adapter 的执行契约至少覆盖当前 `ExecuteSale` 与 `QueryOperation`；authorize/capture/void/refund 仅作为边界说明，不在第 3 章虚假实现。
- 第 3 章定义 connection 身份和执行边界，第 6 章才定义完整 Capability Catalog、三值 Eligibility 与新鲜度策略。

## 3. 逐章可执行合同

### 第 4 章 构建 Routing Context

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | Router 的决定只能和输入快照一样可靠；Routing Context 必须版本化、来源明确、可校验且不携带支付凭证。 |
| 简单系统 | Payment 直接把不断增长的参数列表传给 `route(...)`，Router 猜测缺失字段和默认值。 |
| 规模/失败压力 | 多上游字段同名异义、金额单位漂移、缺失值被当成零、实时信号过期、决策无法复现，以及 token/PAN 被误传到 Router。 |
| 自然推导组件或不变量 | Payment 边界内的 `ContextBuilder`；`RoutingContextV1` 按 Merchant/Transaction/PaymentMethod/RiskReference/RuntimeReference 分组；每个派生或实时字段带来源、`observed_at`/版本。Context 中只允许非敏感类别与派生属性。 |
| 明确留给后章 | Rule 求值留第 5 章；完整 Capability 留第 6 章；Risk Decision 语义留第 9 章；不可变 Decision Store 留第 14 章；Data Residency 留第 19 章。 |
| 窄而完整的 Go 纵向增量 | 新增 `services/payment/internal/routingcontext` 的 Builder/Validator；扩展 `clients.RoutingRequest`、Router HTTP DTO 与 OpenAPI，携带 `schema_version`、merchant country、amount/currency、payment-method category、capture mode 和带时效的非敏感派生字段。Payment 从 Intent 与商户配置构建一次快照，Router 严格解码并拒绝未知/不兼容 schema。 |
| 证明性失败测试 | 构造包含 `payment_method_token`、`pan` 或 `cvv` 的 Router payload 必须在契约/HTTP 层失败；缺失 amount currency 语义或传入过期必需引用时不得默默选路；同一规范化 Context 的序列化摘要必须稳定。 |
| 官方一手资料类别 | PCI SSC 的 CHD/SAD 与 Tokenization 指南；ISO 4217 货币小数位资料；目标 Provider/Payment Method 的官方字段与认证要求。 |
| 依赖前章 | 第 1 章金额与生命周期术语；第 3 章 Router API、connection 身份和敏感数据边界。 |
| 图形锚点 | Mermaid flowchart：五类上游→ContextBuilder→Router；表格：字段/来源/时效/缺失策略/敏感级别；Ian 插图可用“把杂乱包裹过安检再装入透明箱”的隐喻，不画字段精确关系。 |

### 第 5 章 Rule-Based Routing

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | Routing Rule 是不可变、确定、可解释的 Policy artifact；它表达显式业务意图，但不能替代 Provider 资格验证。 |
| 简单系统 | 在 Router 的 Go 代码中不断增加按 merchant/country/currency 的 `if/else` 和紧急开关。 |
| 规模/失败压力 | 例外相互覆盖、优先级不明、同一规则在实例间漂移、紧急改路无法审计，以及 force 到不存在/不合格 connection。 |
| 自然推导组件或不变量 | `RuleSet{Version, EffectiveAt, Rules}`、确定性 Evaluator、稳定优先级与冲突语义；求值输出 Rule Hit/Action/Reason，而不是直接执行 Provider 调用。已发布 artifact 不可变。 |
| 明确留给后章 | 完整 Capability 与“force 也不能越过 Eligibility”的闭环留第 6 章；权重评分留第 9 章；审批、灰度、分发和回滚控制平面留第 20 章。 |
| 窄而完整的 Go 纵向增量 | 在 `services/router/internal/rules` 实现受限 AST（等值、集合、范围）及 `prefer/exclude/force/split` 的确定性求值；Router 启动时从严格 JSON artifact 加载一个不可变版本，Decision 返回 `rule_set_version` 和命中原因。现有 static provider order 成为无规则命中时的安全默认。 |
| 证明性失败测试 | 两条同 scope/priority 且动作冲突的规则使 artifact 校验失败；规则文件顺序变化不改变决定；caller 试图在请求中注入规则或版本必须被 OpenAPI/HTTP strict decoder 拒绝。 |
| 官方一手资料类别 | JSON Schema 官方规范；Go `encoding/json` 与排序/时间官方文档；Merchant/Provider 合同中可被编码为 Policy 的原始条款。 |
| 依赖前章 | 第 4 章稳定 Context 字段与缺失语义；第 3 章 Candidate/connection 身份。 |
| 图形锚点 | Mermaid sequence：加载 artifact→校验→原子启用→求值；表格：动作、可覆盖范围、冲突规则、原因码；Ian 插图可用“贴满便签的岔路口失控后换成编号信号灯”。 |

### 第 6 章 Provider Eligibility

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 先证明一条具体 Provider connection 能合法且技术上处理当前操作，再谈偏好；Eligibility 是带证据的硬过滤，不是分数。 |
| 简单系统 | 把 PSP 名称当 Candidate，并默认一家 PSP 的所有账户、Region、操作和 Payment Method 能力相同。 |
| 规模/失败压力 | 同一 PSP 多 Merchant Account/Connector 能力不同；币种、国家、capture mode、3DS、token 可移植性或合同状态不匹配；能力数据缺失或过期却被当成支持。 |
| 自然推导组件或不变量 | Adapter 发布规范化 `CapabilityDescriptor`；Router 拥有版本化 `CapabilityCatalog` 视图；`EligibilityResult` 为 `eligible/ineligible/unknown` 并带稳定原因码。`unknown` 不默认升级为 eligible。 |
| 明确留给后章 | 费用留第 7 章；经验成功率留第 8 章；Risk guardrail 留第 9 章；动态 Health 只在第 13 章进入 Eligibility；配置审批留第 20 章。Provider Adapter 的执行边界已在第 3 章完成，不另开新章。 |
| 窄而完整的 Go 纵向增量 | 在第 3 章已贯穿的 `RouteCandidate/ProviderConnection` 身份上，新增 `services/router/internal/capability` Catalog/Filter，支持 currency、country、payment method、operation、capture mode、enabled/effective interval；Router contract 返回 catalog version 和三值过滤证据。Payment/Gateway 继续使用第 3 章已经固定的 connection，不在本章重新设计身份。 |
| 证明性失败测试 | 配置同一 PSP 的两个 connection，其中只有一个支持 USD automatic-capture；即使规则强烈 prefer 另一个，也只能选择合格 connection。过期或字段未知的 capability 必须产生 `capability_unknown/stale`，不得调用 Gateway。 |
| 官方一手资料类别 | 各 Provider 官方 API/能力矩阵、API 版本、幂等和 webhook 文档；卡组织/本地 Payment Method 官方 operation 与认证规则；Merchant Account 合同约束。 |
| 依赖前章 | 第 3 章 connection 与 Adapter 边界；第 4 章 operation/capture/context；第 5 章只产生 Policy 偏好。 |
| 图形锚点 | Mermaid flowchart：Catalog→逐项硬过滤→Eligible Set→Policy order；表格：Candidate identity 各字段与所有者、三值结果和降级动作；Ian 插图可用“同品牌多扇门，钥匙、地区和营业牌照各不相同”。 |

### 第 7 章 Cost-Aware Routing

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 在线路由比较的是可解释的 Expected Cost，而不是一项标价；所有货币和费率计算必须确定、可追溯且不使用浮点数。 |
| 简单系统 | 给每家 Provider 配一个 percentage fee，选择百分比最低者。 |
| 规模/失败压力 | fixed fee、cross-border、FX、币种换算和失败尝试改变真实成本；费率有效期或合同版本错配；最低 attempt 成本不等于最低成功支付成本。 |
| 自然推导组件或不变量 | 版本化 `FeeSchedule`、整数 micro-unit/定点数 `CostEstimate`、逐项 `CostBreakdown` 与缺失/过期标志；只对第 6 章 eligible Candidate 估算。Expected Cost 与 Actual Cost 永久分列。 |
| 明确留给后章 | 成功概率留第 8 章；多目标组合留第 9 章；账单、阶梯价最终校准和 Actual Cost 留第 18 章；配置治理留第 20 章。 |
| 窄而完整的 Go 纵向增量 | 新增 `services/router/internal/cost`，支持 fixed+bps+显式 cross-border surcharge 的一条定点数估算链；从不可变 artifact 加载 `fee_schedule_version`，Decision 返回分解与币种。缺少可靠 FX 或费率时由 Policy 明确 fail-closed 或退出成本比较，不能当作零成本。 |
| 证明性失败测试 | 低百分比但高固定费的 Candidate 在小额交易中不应被误选；费率边界 `effective_from/to` 交界结果确定；大金额乘法无溢出，缺失 FX 不得产生零成本或 NaN。 |
| 官方一手资料类别 | Provider/Acquirer 官方费率与账单字段；Card Scheme 官方费用说明；ISO 4217 与受认可 FX 基准发布方的原始资料。 |
| 依赖前章 | 第 4 章金额/币种/国家；第 6 章 eligible Candidate 与 Merchant Account/Region。 |
| 图形锚点 | Mermaid flowchart：eligible set→fee lookup→cost breakdown；表格：费用项/触发条件/单位/版本/在线是否可知；Ian 插图可用“便宜门票后面拖着多张隐藏小票”。 |

### 第 8 章 Success-Rate Routing

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 可用于单笔路由的是带样本量、窗口、置信度和回退层级的条件成功率估计，不是 Provider 总榜。 |
| 简单系统 | 按 Provider 汇总最近一天 Authorization Rate，选择最高者。 |
| 规模/失败压力 | 流量构成和历史路由造成选择偏差；高维分桶稀疏、冷启动、延迟事件、低流量 100% 样本和时间漂移让排行抖动。 |
| 自然推导组件或不变量 | 不可变 `PerformanceSnapshot`、分层 Segment Key、样本量/窗口/`observed_at`、Bayesian smoothing 或显式 shrinkage、从细到粗的回退链。估计值必须携带证据，不能伪装成因果效果。 |
| 明确留给后章 | Cost/Success 综合评分留第 9 章；Health 的分钟级运行信号留第 13 章；指标生产与 SLO 留第 15 章；因果实验留第 16 章；最终结果校正留第 17、18 章；ML 反事实问题留第 21 章。 |
| 窄而完整的 Go 纵向增量 | 新增 `services/router/internal/performance`，读取一个版本化只读 snapshot artifact，按 `candidate_id + connection_revision + country_field + payment_method_context.network + currency` 以及明确的 cohort/denominator contract 查找并逐级回退；输出 `estimate/sample_size/window/confidence/source_segment/version`。先不引入流处理平台。 |
| 证明性失败测试 | 1/1 的 100% 小样本不能无条件击败 900/1000 的稳定样本；细分桶缺失时必须按固定层级回退；过期 snapshot 触发显式降级，候选配置顺序变化不改变结果。 |
| 官方一手资料类别 | Provider/Acquirer 官方交易结果数据字典；Card Network response/advice code 规范；所采用 smoothing、衰减或置信方法的原始论文/规范。 |
| 依赖前章 | 第 4 章稳定分段字段；第 6 章 Candidate identity；第 7 章保持 Expected Cost 与 success estimate 分离。 |
| 图形锚点 | Mermaid tree：细分桶→逐级回退→全局 prior；表格：窗口/样本/估计/置信/新鲜度；Ian 插图可用“只看总成绩榜，错过不同赛道表现”的隐喻。 |

### 第 9 章 Provider Scoring 与 Risk-Aware Routing

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 生产选择是有固定阶段的受约束优化：Eligibility 与 Risk Guardrail 先限定允许集合，Scoring 只在集合内比较可权衡目标，并输出完整分解。 |
| 简单系统 | 把 success、cost、latency、risk 全部塞进一个 weighted sum，最高分胜出。 |
| 规模/失败压力 | 量纲变化、缺失或陈旧 Feature、权重补偿硬约束、同分抖动；Router 重做 Fraud 判断或 Fallback 绕过中央 Risk Block。 |
| 自然推导组件或不变量 | 固定流水线 `Context validation → Eligibility → Rule guardrail → Risk contract → Feature lookup → Score → Stable selection`；`RiskDecision{Action, Reasons, Version, ExpiresAt}` 由 Fraud Engine 拥有；`ScoreBreakdown`、稳定 tie-break 与降级策略由 Router 拥有。 |
| 明确留给后章 | Timeout/执行恢复留第 10～12 章；动态 Health feature 留第 13 章；完整记录留第 14 章；实验留第 16 章；模型留第 21 章。Risk-Aware 不再拥有独立章节。 |
| 窄而完整的 Go 纵向增量 | 新增 `services/router/internal/scoring` 的 typed Feature/Normalizer/Scorer/Selector；扩展 Routing Context 的 Risk Reference/Decision（不得含 Fraud 原始特征），组合第 7、8 章输出并返回 `score_breakdown`、reason codes、feature versions。实现 `block/allow` 与确定性 tie-break；`require_3ds` 只转成第 6 章的硬能力要求，若当前 Adapter slice 不能执行则返回明确的 requires-action/unsupported 结果且不得发起资金操作。 |
| 证明性失败测试 | 一个高分但 capability 不合格的 Candidate 永远不被选；Risk `block` 或过期且配置为 fail-closed 时不调用 Gateway；同分、同版本、同 Context 在实例重启和输入候选乱序后仍选同一 connection；NaN/溢出/缺失 Feature 进入明确降级。 |
| 官方一手资料类别 | Fraud provider 官方 Risk Decision/3DS API；EMVCo 3-D Secure 规范；欧盟官方 PSD2/SCA 法规与监管技术标准；成本和成功率的第 7、8 章一手资料。 |
| 依赖前章 | 第 4 章 Context；第 5 章 Policy；第 6 章 Eligibility；第 7、8 章 Feature。 |
| 图形锚点 | Mermaid pipeline：硬约束闸门与软评分分栏；表格：每个信号是 guardrail/objective/diagnostic、缺失策略与版本；Ian 插图可用“先过安检门，再给合格选手打分”。 |

### 第 10 章 Timeout、Retry 与 Unknown State

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 调用方超时描述的是观测缺失，不是 Provider 未执行；恢复动作必须由副作用确定性、原 operation 身份和剩余 deadline 共同决定。 |
| 简单系统 | Provider 请求超时后立即再发一次，或把所有 5xx 都标成失败。 |
| 规模/失败压力 | 请求已越过进程边界、响应丢失、Payment 在发送后崩溃、query 也超时、webhook 迟到/乱序，以及无限 retry storm。 |
| 自然推导组件或不变量 | 统一 `ProviderObservation` 与 `OutcomeClass{success,failure,pending,unknown}`；持久化 `first_dispatched_at/recovery_deadline_at/next_retry_at/recovery_budget`；`RecoveryPlanner` 只返回 query/wait/same-key resubmit/stop。进入 possible side effect 后禁止新 Attempt。 |
| 明确留给后章 | 商户/API/Provider 三层幂等证明留第 11 章；跨 Provider Fallback 留第 12 章；群体性健康和熔断留第 13 章；独立事实对账留第 17 章。 |
| 窄而完整的 Go 纵向增量 | 在现有 `domain.ProviderOperation`、`app.reconciler` 与 Gateway normalization 上增加 typed transport/HTTP/provider classification 和持久化 recovery deadline/budget；`services/payment/internal/recovery` 计算下一动作。Gateway 对 timeout、断连、无效响应和含糊 5xx 一律回 UNKNOWN；只有 Provider 证明 not-found，或本地 durable state 证明请求从未 dispatch 时，才能用原 operation/connection/key/payload 重送。 |
| 证明性失败测试 | Simulator 先落副作用再丢响应，随后重启 Payment worker；系统必须 query 同一 operation 并收敛，数据库仍只有一个 Attempt、一个 Payment Operation、一个 Gateway Operation。query 连续超时只推进 backoff/budget，不生成第二条资金路径。 |
| 官方一手资料类别 | HTTP RFC 关于幂等与状态码的原文；Provider 官方 timeout、status query、idempotency 与 webhook 语义；Go `context`、`net/http` deadline 官方文档。 |
| 依赖前章 | 第 3、6 章固定 connection；第 9 章选路完成后进入执行阶段。当前 UNKNOWN 基线是起点，本章负责把恢复矩阵与 deadline 补成显式契约。 |
| 图形锚点 | Mermaid sequence：Provider 已成功但响应丢失→UNKNOWN→query/webhook；状态图：operation certainty 与恢复动作；表格：故障位置/能否证明无副作用/允许动作。Ian 插图可用“信件已投进黑箱但回执被风吹走”。 |

### 第 11 章 Idempotency 与支付尝试

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 幂等是对“同一业务操作”的持久化身份、参数不变性和并发所有权定义；不同层的 key 与作用域不能混为一谈。 |
| 简单系统 | Redis 缓存 `Idempotency-Key → response`，进程重启或缓存过期后重新执行。 |
| 规模/失败压力 | 相同 key 不同 payload、并发 confirm、事务提交后响应丢失、Provider key 只在某 Merchant Account/Region 有效、重复/乱序 webhook。 |
| 自然推导组件或不变量 | 四个独立身份：merchant key、内部 operation ID、Provider idempotency key、Provider transaction reference；`(merchant, key)` command fingerprint；`(Adapter 声明的外部幂等 namespace, provider key)` 唯一约束；按 Provider event namespace 的 inbox identity。执行仍固定精确 connection/revision，缓存永不成为事实源。 |
| 明确留给后章 | 新 Attempt 的 Fallback 规则留第 12 章；完整状态参考留附录 A；跨 Region key 所有权留第 19 章；Provider 各异的保留期由 Adapter metadata 暴露但不假设全球统一。 |
| 窄而完整的 Go 纵向增量 | 加固 `services/payment/internal/idempotency` 与 SQL：显式 command scope、canonical fingerprint、in-progress resource correlation/replay；Provider key 由 operation identity 生成并和第 3 章 connection identity 一起持久化。补并发安全的 confirm winner、Gateway request hash 与 webhook inbox payload-hash 冲突语义；恢复继续领取原 operation，不因记录“过期”而新建资金路径。 |
| 证明性失败测试 | 20 个并发相同 confirm key 只产生一个 Attempt/Operation 并返回同一结果；相同 key 不同 command/payload 返回 409；Gateway 同 operation ID 但换 amount、token、connection 或 capture mode 必须在调用 Provider 前拒绝。 |
| 官方一手资料类别 | Stripe、Adyen、PayPal 等目标 Provider 的官方 idempotency 作用域/保留期/并发文档；Provider webhook 重投文档；PostgreSQL unique constraint/transaction 官方文档。 |
| 依赖前章 | 第 10 章 UNKNOWN 与同 operation 恢复；第 3、6 章 connection 作用域；现有 `idempotency_records`、request hash、inbox/outbox 是实现起点。 |
| 图形锚点 | Mermaid entity diagram：Intent→Attempt→Operation→Provider reference 与四类 key；表格：身份/生成者/作用域/生命周期/冲突语义；Ian 插图可用“同一张取号单反复递交，只能领到同一件包裹”。 |

### 第 12 章 Fallback Routing

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | Fallback 是在前一 Attempt 已被证明无副作用且 Policy 明确允许后创建的新 Attempt，不是 timeout 后换一家 PSP 重试。 |
| 简单系统 | PSP A 非 2xx 或超时就按候选列表调用 PSP B。 |
| 规模/失败压力 | UNKNOWN 造成重复扣款、soft/hard decline 被误分类、迟到成功与第二次成功冲突、用户交互/3DS 状态无法迁移、候选循环和总 deadline 被耗尽。 |
| 自然推导组件或不变量 | `OutcomeClassification{certainty, retryable_same_route, fallback_allowed, reason}`；由 Router 生成并由 Payment 固定的 `RoutePlan`；Payment-owned `FallbackPlanner` 检查已尝试集合、次数、deadline、Risk guardrail 与原 Attempt 终态。每个 fallback 都是新 Attempt/Operation/key。 |
| 明确留给后章 | 动态 Health 影响“下一笔交易”留第 13 章；完整 Decision Record 留第 14 章；实验组稳定性留第 16 章；Payment Method 特定 fallback 只实现 tokenized automatic-capture card 的一个明确场景。 |
| 窄而完整的 Go 纵向增量 | 扩展 Router response 为有序且带版本的 Route Plan；Payment 在确认事务中持久化 plan。新增 `services/payment/internal/fallback`，只允许一个显式 allowlist 的 pre-dispatch/conclusive technical failure（`SideEffectNone`）从 Candidate A 创建 Attempt 2 到 Candidate B；UNKNOWN、issuer decline 和 review_required 一律阻断。 |
| 证明性失败测试 | A 明确报告未执行且规则允许时恰有两个 Attempt/Operation；A 发生响应丢失时即使 B 健康也始终只有一条资金路径；若 A 的迟到成功与 B 成功并存，保留两条证据并把 Intent 标成 `review_required`，不得 last-write-wins。 |
| 官方一手资料类别 | Provider/Card Network 官方 decline/advice 与 retry guidance；Provider query/idempotency 文档；3DS、钱包和银行转账官方生命周期文档。 |
| 依赖前章 | 第 9 章冻结允许集合/Risk guardrail；第 10 章 certainty；第 11 章 Attempt 与 operation identity。 |
| 图形锚点 | Mermaid state/sequence：retry、query、wait、fallback 四条互斥路径；表格：Outcome × certainty × allowed action；Ian 插图可用“一扇门回声不明时，不能立刻冲向第二扇门”。 |

### 第 13 章 Circuit Breaker 与 Provider Health

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | Provider Health 是按真实故障维度、带时效和置信度的运行状态；它保护新请求的延迟预算，但不能把业务拒绝伪装成技术故障。 |
| 简单系统 | 每家 PSP 一个全局 `healthy bool`，超过错误阈值立即全量切走，恢复后立即全量切回。 |
| 规模/失败压力 | 故障仅影响某 Region/Connector/operation；低流量误判；切流后失去观测；恢复瞬间惊群；一家慢 Provider 耗尽全局连接池和 worker。 |
| 自然推导组件或不变量 | 以 `candidate_id + operation` 为最小故障域的 `HealthSnapshot{state, observed_at, confidence, version}`；`Healthy/Degraded/Unavailable/Recovering` 带迟滞；half-open probe 配额和 per-connection bulkhead。Decline 不计技术健康。 |
| 明确留给后章 | Dashboard/SLO 留第 15 章；实验因果影响留第 16 章；跨 Region 的本地/全局信号组合留第 19 章；配置发布留第 20 章。 |
| 窄而完整的 Go 纵向增量 | 新增 `services/router/internal/health` 状态机与 Router-owned PostgreSQL snapshot/observation store；Gateway 通过自己的 outbox 非阻塞投递有限基数的 technical observation，Router 内部 ingest endpoint 幂等消费。Eligibility 对 `Unavailable` 硬排除，Scoring 对 `Degraded` 加显式惩罚；Gateway 增加 per-connection semaphore/bulkhead。 |
| 证明性失败测试 | `candidate_a/EU/sale` 连续 timeout 只熔断该 connection，不影响同 PSP 的 US connection；issuer decline 不触发熔断；fake clock 下 half-open 只放行限定探测并在失败后重新打开，100 个并发请求不能越过 bulkhead。 |
| 官方一手资料类别 | Provider 官方 status/incident、SLA、rate-limit 与 regional endpoint 文档；Go synchronization/HTTP transport 官方文档；采用的熔断算法原始规范或论文。 |
| 依赖前章 | 第 6 章精确 Candidate；第 9 章 Eligibility/Scoring 插槽；第 10 章技术 outcome 分类；第 12 章区分 per-request fallback 与 future-request health routing。 |
| 图形锚点 | Mermaid state diagram：Healthy↔Degraded↔Unavailable↔Recovering；表格：信号/维度/是否影响 Eligibility/TTL；Ian 插图可用“同一大楼只有一部电梯坏了，不能封整栋楼”。 |

### 第 14 章 Routing Decision Store

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | “为什么走这条 route”必须由不可变事实回答；可审计记录保存当时的输入摘要、候选证据和版本引用，而不是用当前配置重算过去。 |
| 简单系统 | 日志只写 `selected_provider=PSP_B`，配置和指标滚动后再也无法解释。 |
| 规模/失败压力 | 一笔 Intent 多次 Attempt/Fallback、规则与 Feature 已变化、日志采样丢失、敏感 Context 被复制，以及审计存储故障反向阻断支付。 |
| 自然推导组件或不变量 | Router 构造 canonical `DecisionEnvelope` 和 digest；Payment 在创建 Attempt 的同一事务中固定执行快照；append-only Decision Store 是可查询投影，通过 outbox 异步摄取。相同 decision ID 不同 hash 必须 quarantine。 |
| 明确留给后章 | Metric/Trace 留第 15 章；Experiment assignment 留第 16 章；差异修复留第 17 章；保留期/区域复制留第 19 章；谁能发布配置留第 20 章。Decision Store 不是 Payment Ledger。 |
| 窄而完整的 Go 纵向增量 | 扩展 Router Decision 为包含 context digest、全部 Candidate/exclusion、rule hits、score breakdown 和 capability/fee/performance/health/risk 版本的 envelope；Payment 的 `PrepareConfirmation` 原子保存该 envelope 与 hash 并写 outbox。新增 `services/router/internal/decisionstore` 的 PostgreSQL append-only ingest/query 投影，消费重复事件时按 ID+hash 幂等。 |
| 证明性失败测试 | 决策后修改全部配置，按 decision ID 查询仍得到原选择与理由；Decision Store 停机时 Payment 可凭本地执行快照安全继续且 outbox 积压，恢复后恰好形成一条投影；同 ID 不同 payload 被隔离而非覆盖。 |
| 官方一手资料类别 | PCI SSC 日志与数据最小化要求；Provider 官方 transaction/reference/event schema；PostgreSQL JSONB、constraint 与 partitioning 官方文档。 |
| 依赖前章 | 第 4～9 章提供 Context/规则/资格/Feature/分数版本；第 10～13 章提供 operation、fallback 与 health 证据；第 11 章提供稳定身份。 |
| 图形锚点 | Mermaid flowchart：Router envelope→Payment transaction/outbox→Decision Store；表格：记录字段/事实还是引用/所有者/敏感级别；Ian 插图可用“事故后打开封存的黑匣子”。 |

### 第 15 章 Routing Observability

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 可观测性必须分开回答 Router 自身是否健康、Provider 调用发生了什么、支付结果如何、经济结果何时成熟；总体 Authorization Rate 不能替代这些问题。 |
| 简单系统 | 一个全局成功率 Dashboard 和自由文本日志。 |
| 规模/失败压力 | Simpson's Paradox、流量结构变化、高基数 label 爆炸、异步结果延迟、Router latency 与 Provider latency 混淆、告警没有负责人或动作。 |
| 自然推导组件或不变量 | 分层 SLI/metric taxonomy；受控维度 registry；Intent/Attempt/Decision/Operation ID 只进 trace/log correlation 不进 metric label；Outcome 的 observation time 与 effective time 分离；每个告警绑定 Runbook。 |
| 明确留给后章 | 实验分组与因果判断留第 16 章；最终事实校正留第 17 章；实际费用留第 18 章；100k TPS 容量验证留第 22 章和附录 C。 |
| 窄而完整的 Go 纵向增量 | 新增共享但按服务初始化的 `internal/telemetry` 契约：Payment/Router/Gateway 传播 `traceparent` 与 request correlation；暴露 decision latency/error/no-candidate/stale-feature、gateway technical outcome、UNKNOWN age、outbox depth 等有限 label 指标；OpenTelemetry exporter 可选、无 endpoint 时 no-op。 |
| 证明性失败测试 | 运行 success、decline、timeout 三场景后 scrape `/metrics`，能区分 Router 决策错误与 Provider timeout；输出不得出现 merchant ID、Intent/Attempt ID、token、PAN/CVV；高基数输入不会增加 label cardinality。 |
| 官方一手资料类别 | OpenTelemetry semantic conventions/specification；Prometheus exposition 与 instrumentation 官方文档；Provider 官方 outcome/latency/status 定义；支付结果口径的 Card Network/Acquirer 数据字典。 |
| 依赖前章 | 第 14 章关联证据；第 10、13 章 timeout/health 分类；第 7、8 章成本与成功信号的不同成熟时间。 |
| 图形锚点 | Mermaid layered diagram：service SLI→provider operation→payment outcome→economic outcome；表格：metric/口径/label/成熟延迟/负责人；Ian 插图可用“一个总温度计掩盖不同房间着火”。 |

### 第 16 章 Routing Experimentation

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 支付路由实验先固定业务单元、版本和安全护栏，再逐级从 shadow 到小流量；“随机切 10%”本身不是可归因实验。 |
| 简单系统 | 每次请求随机 90/10 选择新旧策略，并只看即时成功率。 |
| 规模/失败压力 | 同一 Intent 的 retry/fallback 跨组、样本比例不匹配、Provider 容量互相干扰、配置中途变化、欺诈/拒付/成本反馈延迟，以及实验策略触发真实副作用。 |
| 自然推导组件或不变量 | 以 Payment Intent 为默认稳定分桶单位；`ExperimentAssignment{experiment,variant,version,bucket}` 固定在 Decision；shadow 只产 Decision 不产 Provider Operation；Primary/Guardrail metric、停止条件和 kill switch 显式化。 |
| 明确留给后章 | 数据纠错和最终结果留第 17、18 章；跨 Region 分桶一致性留第 19 章；实验配置审批留第 20 章；bandit/ML 探索留第 21 章。 |
| 窄而完整的 Go 纵向增量 | 新增 `services/router/internal/experiment` 的稳定 hash assigner、版本化 experiment artifact 和 shadow evaluator；Router response/Decision Record 带 assignment 与 shadow outcome，Payment 只执行 champion route。实现可热禁用但保持已分配 Intent 稳定的 kill switch。 |
| 证明性失败测试 | 同一 Intent 的重复 confirm、query 和 fallback 永远属于同一 variant；1000 个固定 Intent 的分桶快照可重复且候选顺序无关；shadow Candidate 即使得分最高也不会出现在 Gateway Operation；kill switch 后无新 treatment assignment。 |
| 官方一手资料类别 | 所采用 A/B 检验、SRM 与 sequential stopping 的原始论文/规范；Provider 官方容量/限流与 sandbox 文档；隐私监管方关于实验数据用途的原文。 |
| 依赖前章 | 第 14 章不可变 Decision Record；第 15 章可靠指标；第 9 章可插拔 scorer；第 12 章 Route Plan/Intent 稳定性。 |
| 图形锚点 | Mermaid progression：offline replay→shadow→1%→10%→champion；表格：阶段/是否有副作用/指标成熟期/停止条件；Ian 插图可用“先在影子赛道试跑，再打开一小段真实赛道”。 |

### 第 17 章 Reconciliation

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | Reconciliation 是多源事实的持久化摄取、匹配、差异分类和受控修复；它补在线链路的盲区，但没有权力用一条迟到记录覆盖既有资金事实。 |
| 简单系统 | 相信 webhook 最终一定到达，偶尔用 SQL 把 `processing` 批量改成 `succeeded`。 |
| 规模/失败压力 | webhook 丢失/重复/乱序、Provider 人工操作、query 与内部终态冲突、reference 碰撞、回看窗口与 checkpoint 崩溃重跑。 |
| 自然推导组件或不变量 | 独立 deployable reconciliation worker 但仍服从 Payment 数据所有权；`Run/Checkpoint/Evidence/Difference/Resolution` 持久化；Matching 先强 identity 再受控复合键；自动修复只允许单调且无冲突的转换，其余进入 `review_required`。 |
| 明确留给后章 | Settlement line 与实际费用留第 18 章；跨 Region reconciliation ownership 留第 19 章；Ledger adjustment 由外部 Ledger/人工工作流拥有，不在 Router 或 reconciler 内伪造。 |
| 窄而完整的 Go 纵向增量 | 把当前 `app.RunReconciler` 明确保留为在线 UNKNOWN recovery；新增 `services/payment/cmd/reconciler` 与 `internal/reconciliation`，持久化扫描 checkpoint，并通过 Gateway query/归一化 provider extract 对 aged operation 做匹配和 `missing/status/reference` 差异分类。复用 Payment `ApplyOutcome/RequireManualReview`，不直接 SQL 改终态。 |
| 证明性失败测试 | 内部 `failed` 与 Provider `succeeded` 冲突时保存双方证据并 `review_required`，不得覆盖；丢失 webhook 但 Provider 查询证明 captured 时安全收敛；worker 在 checkpoint 提交前崩溃并重跑不重复修复、不丢差异。 |
| 官方一手资料类别 | Provider 官方 transaction search、webhook delivery、reconciliation/transaction report 文档；Card Network/Acquirer 官方 reference 与 clearing record 定义；PostgreSQL transaction/locking 官方文档。 |
| 依赖前章 | 第 10 章 query/UNKNOWN；第 11 章 identity/inbox；第 14 章 decision correlation；第 15 章差异与积压观测。 |
| 图形锚点 | Mermaid flowchart：ingest→normalize→match→classify→auto repair/manual review；表格：差异类型/证据/自动动作/人工出口；Ian 插图可用“两本账摊开逐行对照，矛盾处贴红签”。 |

### 第 18 章 Settlement 与 Cost Attribution

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 路由的经济结果只能从结算/费用事实回连到原 Decision、Attempt 和 connection 后计算；Expected、Accrued、Actual Cost 必须分层。 |
| 简单系统 | Authorization 成功时按在线报价立即记作最终处理成本。 |
| 规模/失败压力 | 结算文件延迟、FX 与多币种、scheme/cross-border/refund/chargeback fee、多个 Attempt、重复或更正行、无法匹配的 Provider reference。 |
| 自然推导组件或不变量 | 独立 `SettlementFile/SettlementLine/CostAttribution` 事实模型；原始文件只追加、line identity 幂等、更正通过新事实表达；使用整数/定点数并保存原币种、换算率与来源版本。Router 的 Fee Model 只消费校准制品，不写结算事实。 |
| 明确留给后章 | 跨 Region 数据驻留留第 19 章；自动发布新费率留第 20 章；模型使用 Actual Cost 留第 21 章。Ledger、Payout 与会计分录保持外部边界。当前代码仍只承诺 automatic-capture SALE；manual Capture/Void/Refund API 未实现时必须明示，不能因能摄取相关报表行而宣称支持。 |
| 窄而完整的 Go 纵向增量 | 新增 `services/settlement` worker/CLI 与独立 PostgreSQL schema，先实现一种版本化 CSV/JSON provider report adapter：幂等摄取 captured sale 与 fee line，通过 `provider connection + provider reference + amount/currency` 关联 Operation/Decision 事件，产出 Expected-vs-Actual calibration report。服务间通过 API/事件快照，不跨库读表。 |
| 证明性失败测试 | 同一文件/line 重投不重复计费；无法匹配或 amount/currency 冲突进入 exception queue；多 Attempt 只把成功资金与各 attempt fee 分别归因，成本分解之和必须精确等于原始 line 总额。 |
| 官方一手资料类别 | Provider/Acquirer 官方 settlement、balance、fee 与 dispute report schema；Card Scheme 官方 clearing/fee 规则；ISO 4217 与官方 FX/汇率来源。 |
| 依赖前章 | 第 7 章 Expected Cost；第 11 章稳定 operation/reference；第 14 章 Decision 关联；第 17 章可信差异与最终状态。 |
| 图形锚点 | Mermaid lineage：Decision→Attempt→Operation→Provider ref→Settlement line→Actual cost；表格：Expected/Accrued/Actual 的来源和成熟期；Ian 插图可用“几天后才寄到的一叠账单揭示真实票价”。 |

### 第 19 章 Multi-Region Payment Routing

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 多 Region 的难点不是复制无状态 Router，而是确保一个 Intent/Operation 在任一时刻只有一个有权产生资金副作用的 owner，同时让决策数据在合规边界内本地可用。 |
| 简单系统 | 把 Payment、Router 和数据库复制到三地，用 DNS active-active 分流。 |
| 规模/失败压力 | 跨区数据库延迟、重复 merchant key、Intent split-brain、配置版本漂移、Region failover 时 UNKNOWN 未清、Provider endpoint/idempotency scope 按 Region 不同、数据驻留禁止复制。 |
| 自然推导组件或不变量 | `HomeRegion`、Intent ownership lease 与单调 fencing epoch；所有创建/发送 Operation 的事务验证 epoch；Router 区域自治并记录本地 config/feature version；全局身份只保存最小必要索引。Failover 是显式状态迁移，不是 DNS 自动放开写权限。 |
| 明确留给后章 | 配置审批/分发控制平面留第 20 章；模型全球发布留第 21 章；100k TPS 完整拓扑留第 22 章。具体法律适用性由合规/法律团队决定，不写入 Router 猜测。 |
| 窄而完整的 Go 纵向增量 | 为 Payment 增加 `region_id/home_region/ownership_epoch` 与 `RegionOwnership` 接口；在同一 PostgreSQL 测试环境启动两个逻辑 Region 实例，使用全局最小索引和 fencing transaction 保证 merchant idempotency/Intent owner 唯一。Router/Gateway request 带 execution region，Operation 继续固定原 connection 与 epoch；提供受控 failover command。 |
| 证明性失败测试 | 两个 Region 同时 confirm 同一 Intent，只有当前 epoch 能创建/发送 Operation；旧 owner 在 lease 过期后恢复仍被 fencing 拒绝。UNKNOWN 未解决时 failover 不创建新资金路径；配置版本暂时不同仍可从 Decision 解释各自选择。 |
| 官方一手资料类别 | 目标云厂商官方 Region/故障域/数据驻留文档；监管机构关于跨境数据与保存要求的原文；Provider regional endpoint、账户和 idempotency scope 官方文档；PostgreSQL consistency/replication 官方文档。 |
| 依赖前章 | 第 11 章 global business identity；第 10、12 章 UNKNOWN/Fallback 安全；第 13 章局部 Health；第 14 章版本快照；第 17 章恢复所有权。 |
| 图形锚点 | Mermaid deployment：三个区域 data plane、全局最小索引、异步反馈；sequence：owner fencing 与灾难切换；表格：数据集/一致性/RPO/RTO/驻留范围。Ian 插图可用“多座厨房共用一张唯一取餐号，只有持当前令牌者能出餐”。 |

### 第 20 章 Routing Configuration Platform

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 配置是可造成全局资金影响的生产制品；必须经过 schema/语义/影响验证、职责分离、不可变发布和可证明回滚，且控制平面故障不能阻断在线支付。 |
| 简单系统 | 运营人员直接 UPDATE Router 数据库或替换实例环境变量。 |
| 规模/失败压力 | 并发编辑、force 到无能力 connection、权重和费率单位错误、部分 Region 收到半套版本、发布人自批、控制平面停机、无法重现/回滚。 |
| 自然推导组件或不变量 | `Draft → Validated → InReview → Approved → Published → Superseded/RolledBack`；作者与审批者分离；不可变 `ConfigBundle` 把 Rule/Capability/Fee/Experiment/Model 引用固定为一个 manifest+digest；Router 原子切换并保留 last-known-good。动态 Health 不属于 bundle。 |
| 明确留给后章 | 模型训练与 online scoring 留第 21 章；综合容量与组织 Runbook 留第 22 章/附录 C。Break-glass 仍需事后审计，不能成为绕过不可变版本的普通接口。 |
| 窄而完整的 Go 纵向增量 | 新增 `services/routing-config`（独立 binary/schema）的最小 API 与领域状态机：创建 Config bundle、运行现有各章 validator、不同 principal 审批、发布带 digest 的 manifest、Router watcher 原子 apply/ack、按上一已发布版本 rollback。所有动作写 append-only audit event。 |
| 证明性失败测试 | artifact 作者不能批准自己的发布；引用不存在/不合格 Candidate 或费率溢出的 bundle 无法进入 Approved；分发中断时 Router 要么使用完整旧 bundle 要么完整新 bundle，绝不混用；控制平面停机仍以 last-known-good 决策。 |
| 官方一手资料类别 | NIST 关于 RBAC、least privilege、separation of duties 与 audit 的标准；SLSA/provenance 官方规范；目标 Provider API/capability/fee 版本原文；部署平台原子 rollout/rollback 官方文档。 |
| 依赖前章 | 第 5～9 章 artifact validator；第 14～16 章回放/审计/实验；第 19 章多 Region version convergence。 |
| 图形锚点 | Mermaid state diagram：配置生命周期；sequence：publish manifest→各 Region ack→rollback；表格：artifact 类型/owner/validator/发布频率/降级策略；Ian 插图可用“双钥匙才能打开全局控制台”。 |

### 第 21 章 Intelligent Routing

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 模型只是 Scoring Pipeline 的可替换估计器；只有在可识别偏差、离线证据和安全探索成立时才上线，并永远受 Eligibility、Risk、容量和回滚护栏约束。 |
| 简单系统 | 用历史已选 route 的结果训练“成功预测模型”，直接让预测最高者接管全部流量。 |
| 规模/失败压力 | Logged Policy/selection bias、缺少未选 Candidate 反事实、延迟欺诈/拒付标签、概念漂移、稀疏冷启动、inference timeout/NaN、探索会触及真实资金。 |
| 自然推导组件或不变量 | 版本化 `ModelArtifact{feature_schema,training_window,metrics,digest}`、logged propensity、离线 OPE、shadow/challenger、受限 exploration budget；`ModelScorer` 实现第 9 章统一接口，失败时回退已验证 rule/static scorer。模型不拥有 Candidate generation。 |
| 明确留给后章 | 不实现自治 RL 或无护栏全量探索；组织级上线与容量验证留第 22 章。若数据/收益条件不成立，“保持 Rule-Based”是有效交付结论。 |
| 窄而完整的 Go 纵向增量 | 新增 `services/router/internal/model` 的严格 artifact loader 与确定性线性/logistic inference（定界 Feature、拒绝 NaN/Inf），以及 `tools/routing-replay` 用 Decision Store 数据计算覆盖率和一种带 propensity 的离线估计。先 shadow，经第 16/20 章门禁后才能成为 scorer；Decision 固定 model/feature version。 |
| 证明性失败测试 | 模型给 ineligible Candidate 最高分仍不能越过第 6 章；inference timeout、未知 feature schema、NaN/过期 model 立即回退 last-known-good scorer；缺少 propensity/support 的 replay 必须拒绝输出“模型更优”的结论。 |
| 官方一手资料类别 | Contextual Bandit、off-policy evaluation 与所用算法的原始论文；模型格式/数值库官方规范；Provider/Card Network 官方 outcome 数据字典；适用监管机构的 automated decision/数据用途要求。 |
| 依赖前章 | 第 8 章偏差与统计信号；第 9 章 scorer interface/guardrail；第 14～16 章日志、指标、shadow/实验；第 20 章 artifact 审批/回滚。 |
| 图形锚点 | Mermaid lifecycle：log→train→OPE→shadow→experiment→serve→rollback；表格：方法/所需数据/探索风险/可解释性/降级；Ian 插图可用“模型驾驶员坐副驾，安全护栏和刹车仍由系统掌握”。 |

### 第 22 章 从零设计 Global Payment Routing Platform

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 综合架构的价值在于用一组明确 SLO、容量和故障场景检验前 21 章的不变量，而不是把所有组件画进一张大图。 |
| 简单系统 | 单 Region、单 PostgreSQL、静态 Policy 与少量 Provider connection，先跑通可审计 automatic-capture card。 |
| 规模/失败压力 | 10,000 Merchant、50 国家、30 币种、20 Payment Method、15 PSP、100k TPS 峰值；Provider timeout、错误配置、Region 故障、反馈延迟和团队所有权冲突。 |
| 自然推导组件或不变量 | 明确 functional/non-functional requirement、容量预算和非目标；按阶段演进的 data/control plane；端到端 deadline、故障域与降级顺序；四条系统级验收旅程。所有复杂组件都能追溯到前章压力。 |
| 明确留给后章 | 无后续机制章；尚未实现的 live PSP connector、PCI certification、全部 Payment Method、真实多云 100k TPS 必须列为外部工程/合规工作，不用 Simulator 或设计数字冒充完成。 |
| 窄而完整的 Go 纵向增量 | 不新增新的业务抽象；新增 `test/system` 场景编排、fault injection 与 `tools/capacity-plan`，组合现有 Compose/服务验证正常支付、Provider response loss、坏配置回滚和 Region owner failover。输出实测吞吐/延迟/资源与模型假设，扩展 Makefile/CI 为全书总回归门禁。 |
| 证明性失败测试 | 四条旅程必须分别证明：正常成功；timeout 后仅一条资金路径并收敛；无效 bundle 不进入 data plane/可回滚；旧 Region fencing 后不能执行。压测必须统计 HTTP 非 2xx、UNKNOWN backlog、outbox age 和资源指标，不能只报成功 RPS。 |
| 官方一手资料类别 | 前 4～21 章已核验的一手资料；目标云/数据库/部署平台容量和故障域官方文档；Provider rate limit/SLA/sandbox/live readiness 文档；适用支付与安全标准原文。 |
| 依赖前章 | 第 4～21 章全部，但只做组合、权衡和验证，不重新发明规则、幂等、健康、配置或模型。 |
| 图形锚点 | Mermaid C4-like deployment（控制在一页）和四张关键 sequence；表格：SLO/预算/降级/owner、容量公式与测量证据；Ian 插图可用“把前章零件装成一台可接受故障演练的机器”。 |

## 4. 附录可执行合同

### 附录 A Payment Intent、Attempt 与 Provider Request 状态模型（[交付记录](appendix-a-payment-state-models.md)）

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | Intent、Attempt 与 Provider Operation 是不同聚合层次；把它们压成一个 `payment_status` 会让 UNKNOWN、Fallback、Capture/Refund 和冲突事实无法正确表达。 |
| 简单系统 | `created/succeeded/failed` 一列同时表示业务目标、一次调用和 Provider 资金状态。 |
| 规模/失败压力 | 一 Intent 多 Attempt、一 Attempt 多 operation、授权与请款分离、重复/乱序事件、终态冲突、部分金额，以及各 Payment Method 生命周期不同。 |
| 自然推导组件或不变量 | 三层状态机与基数；单调转换、terminal protection、optimistic version、active/winning Attempt 约束、operation certainty；“Provider Request”在正文泛指外部请求，代码权威名统一为 `ProviderOperation`。 |
| 明确留给后章 | 不强迫卡、钱包、银行转账共用一张状态表；未实现的 manual capture/void/refund 只给参考转换与 Adapter 约束，不得标成当前 API 能力。Ledger 状态不并入 Intent。 |
| 窄而完整的 Go 纵向增量 | 不新增服务；把 `services/payment/internal/domain` 的 enum、transition 和 invariant checker 整理为可执行参考，增加 table-driven `state_spec_test.go` 与 machine-readable transition fixture，供附录 Mermaid/表格和测试共同使用。迁移约束与 Go invariant 一一对应。 |
| 证明性失败测试 | 已 captured 的 Operation 收到 conclusive failure 不得回退为 failed，而应保留成功事实并进入 `review_required`；UNKNOWN Attempt 不得创建第二次 confirm；winning Attempt 与 captured amount/reference 不一致时数据库约束拒绝提交。 |
| 官方一手资料类别 | Provider 官方 Payment/Authorization/Capture/Void/Refund 状态文档；Card Network 与本地 Payment Method 生命周期规范；PostgreSQL constraint/transaction 官方资料。 |
| 依赖前章 | 第 1 章领域入口；第 10～12 章 UNKNOWN/幂等/Fallback；第 17 章冲突证据。附录负责统一参考，不重复章节推导。 |
| 图形锚点 | 三张 Mermaid state diagram（Intent/Attempt/Operation）与一张基数图；表格：事件→允许前态→后态→资金证据→失败动作；Ian 插图不建议使用，精确状态图信息增益更高。 |

### 附录 B 安全、合规与数据边界（[交付记录](appendix-b-security-compliance.md)）

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | Router 应只持有完成决策所需的最小数据；Tokenization、加密或微服务拆分都不能自动消除 PCI/隐私/密钥管理责任。 |
| 简单系统 | 所有支付字段进入统一 Context、日志、事件和分析仓，服务共享 Provider secret 与数据库账号。 |
| 规模/失败压力 | PAN/CVV 泄漏到 trace、token 可跨 Provider 性被误判、webhook 伪造、配置投毒、过宽 service role、secret 长期不轮换、跨区复制违反目的/驻留限制。 |
| 自然推导组件或不变量 | 数据分类与 trust-boundary map；Router/Decision forbidden-field allowlist；service identity 与最小权限；Provider credential 按 connection 隔离；webhook 验签+重放防护；保留/删除和审计策略。CVV/SAD 授权后不得保存。 |
| 明确留给后章 | 本附录不提供法律结论、不宣称 PCI certification，也不实现 HSM/KMS 产品；具体市场要求由安全、合规和法律团队确认。业务 Risk Decision 仍属于第 9 章，不在这里重做 Fraud Engine。 |
| 窄而完整的 Go 纵向增量 | 新增可复用的安全边界测试/工具：Router/Decision JSON allowlist 与 forbidden-key scanner、结构化日志 redaction handler、production config fail-closed、Gateway webhook replay/signature 测试；为 Payment/Gateway/Router/Config/Settlement 定义独立 DB role migration/grant 示例和 secret provider interface。 |
| 证明性失败测试 | PAN/CVV/SAD 字段无论大小写或嵌套都不能通过 Router/Decision contract；日志与 outbox fixture 不含 token/secret；伪造或重放冲突 webhook 被拒绝/隔离；production 使用 simulator、默认 token 或共享高权限 DB role 时 launch check 失败。 |
| 官方一手资料类别 | PCI SSC PCI DSS 与 Tokenization/SAD 指南；EMVCo Tokenization/3DS 规范；NIST key management、identity 与 access control 标准；监管机构数据保护/驻留原文；Provider webhook security 官方文档。 |
| 依赖前章 | 第 3、4、6 章 boundary/Context/connection；第 14 章 Decision data minimization；第 19、20 章 residency/RBAC。 |
| 图形锚点 | Mermaid trust-boundary/data-flow diagram；表格：数据类别/允许服务/允许落盘/日志策略/保留期；Ian 插图可用“透明管道外有多道闸门，敏感包裹不进入 Router 房间”。 |

### 附录 C 容量规划与上线检查清单（[交付记录](appendix-c-capacity-planning-and-launch-checklist.md)）

| 合同项 | 内容 |
| --- | --- |
| 核心命题 | 容量是由外部 fan-out、失败放大、状态积压和故障余量共同决定的可测预算；一个脱离流量分布与错误率的 TPS 数字没有上线意义。 |
| 简单系统 | 用单次 happy-path benchmark 的平均 RPS 作为生产容量，并只检查进程 CPU。 |
| 规模/失败压力 | retry/query/fallback 放大、Provider rate limit、连接池饥饿、热点 Merchant、cache miss storm、outbox/reconciliation backlog、Region failover 剩余容量和非 2xx 被 benchmark 工具漏报。 |
| 自然推导组件或不变量 | 负载方程 `ingress × (route + execute + retry/query/fallback)`；按 service/dependency 的 latency/queue/resource budget；steady-state、slow Provider、cache loss、Region loss 四类压测；Go/No-Go 证据与 owner/runbook。 |
| 明确留给后章 | 不提供脱离目标部署的“系统最大 TPS”；真实 Provider 限额、真实云资源和 live 交易演练必须由上线环境复测。附录复用第 22 章 harness，不再创建第二套压测系统。 |
| 窄而完整的 Go 纵向增量 | 在第 22 章 `tools/capacity-plan` 基础上增加 `tools/launchcheck`：读取显式参数计算 fan-out/余量，探测 readiness、连接池/timeout 配置、outbox/UNKNOWN age、config version 与 simulator/live 门禁；输出机器可读报告和非零失败码。 |
| 证明性失败测试 | 注入慢 Provider 和 5% timeout 后，报告必须计入 query/recovery 请求、HTTP 非 2xx、queue age 与资源曲线；Region-loss 模式若剩余容量不足或 UNKNOWN 超阈值则 Go/No-Go 为 No-Go，不能只因进程存活而通过。 |
| 官方一手资料类别 | Provider 官方 rate limit/SLA/连接建议；Go runtime、`net/http`、`database/sql`/pgx pool 官方文档；Kubernetes probe/resource/rollout 官方文档；目标云数据库/网络容量官方资料。 |
| 依赖前章 | 第 13、15 章 Health/SLO；第 19 章 Region failover；第 20 章配置版本；第 22 章综合场景与实测 harness。 |
| 图形锚点 | Mermaid queue/fan-out flow；表格：输入假设→公式→容量→实测→余量、上线项/owner/evidence/rollback；Ian 插图可用“高峰水流经过多处分叉和回流，最窄管道决定上限”。 |

## 5. 跨章状态与术语所有权矩阵

| 对象/术语 | 权威所有者与唯一写入边界 | 其他组件可以做什么 | 首次固定/主要扩展章节 | 不可违反的不变量 |
| --- | --- | --- | --- | --- |
| Payment Intent | Payment/Orchestrator；Payment PostgreSQL | API 读取 projection；Reconciler 通过领域命令提交证据 | 第 1、3 章；附录 A | 表达商户业务目标，不等同于一次 Provider 调用；只有 Payment 改状态 |
| Payment Attempt | Payment/Orchestrator | Router 为它生成 Decision；Gateway 只看 attempt ID 作为关联 | 第 3、11、12 章 | Fallback 才可能创建新 Attempt；UNKNOWN 阻断新资金路径 |
| Provider Operation | Payment/Orchestrator 拥有逻辑 operation；Gateway 拥有外部执行副本 | Adapter execute/query；webhook/reconciliation 提交 Observation | 第 3、10、11 章；附录 A | operation ID、connection、payload、Provider key 固定；Capture/Void/Refund 绑定原 reference |
| Merchant idempotency key | Merchant 生成，Payment 定义 `(merchant,key)` 语义并持久化 | API gateway 传递，不得改写 | 第 11、19 章 | 同 key 不同 command/payload 冲突；跨 Region 仍只有一个 winner |
| Provider idempotency key | Payment 从 operation 派生；Gateway/Adapter 按 Provider contract 使用 | Provider 在 Adapter 声明的外部幂等 namespace/期限内去重 | 第 10、11 章 | 不是跨 PSP 全局去重；唯一作用域跟随 Provider 实际外部 namespace，执行仍固定精确 connection/revision |
| Provider transaction/reference | Provider 产生；Gateway 原样保存并归一化关联 | Payment 只在验证 identity 后绑定；Settlement/Reconciliation 匹配 | 第 3、10、17、18 章 | 不可伪造或复用到另一 Attempt；冲突进入 review |
| Provider / Merchant Account / Connector / Region | Connector/配置所有者维护；Gateway 掌握 endpoint/credential | Router 用其只读视图组成 Candidate；Payment 固定所选快照 | 第 3、6 章 | `Provider` 品牌名不能替代可执行 connection identity |
| Provider Adapter | Provider Gateway | Router 只读 capability descriptor；Payment 调统一 Gateway contract | 第 3、6 章 | 只执行已选 connection 并归一化语义，不能暗中路由 |
| Capability Catalog | Connector/Provider 集成团队发布事实；Router 持有决策视图 | Config Platform 验证/分发；Eligibility 读取 | 第 6、20 章 | 版本、effective interval、新鲜度和 unknown 明确；不是 Health |
| Routing Context | Payment 的 ContextBuilder 固定快照；源系统拥有各自原始字段 | Router 严格校验/消费；Decision 只存摘要/必要事实 | 第 4、14 章 | 不含 token/PAN/CVV/SAD；缺失和过期不能伪装成默认真值 |
| Routing Rule/Policy | Router 定义求值语义；Config Platform 管 artifact 生命周期 | Merchant/运营提交意图；Payment 不注入临时规则 | 第 5、20 章 | 已发布版本不可变；Rule 偏好不能越过 Eligibility |
| Eligibility Result | Router | Payment 执行 selected Candidate；Decision Store 审计原因 | 第 6、9、13 章 | 硬约束先于 Scoring；unknown 不默认 eligible |
| Expected Cost / Actual Cost | Router Cost Model 拥有估算；Settlement 服务拥有实际事实 | Config Platform 发布校准 fee artifact | 第 7、18、20 章 | 两者分列且带币种/版本；Router 不改 Settlement line |
| Performance Estimate | Performance snapshot producer 发布；Router 只读估计 | Observability/Reconciliation 提供校正数据 | 第 8、15～18 章 | 带窗口、样本、置信与选择偏差声明；不当作因果优势 |
| Risk Decision | Fraud/Risk Engine | Payment 传递版本化最小 contract；Router 执行 guardrail/有限偏好 | 第 9 章 | Router 不重算 Fraud；block 不可被分数、实验或 fallback 绕过 |
| Routing Decision / `RoutePlanDecision` | Router 构造语义、固定 candidate order/版本/预算/有效期与 digest；Payment 固定执行快照 | Decision Store 保存查询投影；Gateway 只收选定 connection | 第 3、9、12、14 章 | Decision 不产生资金副作用；guardrail-only revalidation 只能 veto，不能新增/重排/重评分/续期；过期必须 Stop |
| `FallbackExecutionBinding` | Payment/Orchestrator 在 successor transaction 中写入 | Router-owned guardrail 检查固定 next；Decision Store 可投影 lineage | 第 12、14、22 章 | 只能引用既有 `RoutePlanDecision`、前序 Attempt、固定 next 与新鲜证据；不是 Router Decision，也不能授权 Plan 外执行 |
| Provider Health | Health subsystem/Router-owned store 聚合状态；Gateway 提供技术 observation | Eligibility/Scoring 消费 snapshot；Observability 展示 | 第 13、15、19 章 | 按故障维度和时效；Decline 不等于技术不健康；不是 Capability |
| Decision Record | Router 构造，Payment outbox 保证已执行快照最终投影；Decision Store append-only | Metrics/Experiment/Replay/Cost attribution 读取 | 第 14～18、21 章 | 相同 ID 不同 hash 隔离；不复制不必要敏感数据；不是 Ledger |
| Experiment Assignment | Router Experiment evaluator；Config Platform 管 artifact | Payment 只执行 champion 并固定 assignment | 第 16、20、21 章 | 同一 Intent 稳定；shadow 不创建 Provider Operation |
| Reconciliation Evidence/Difference | Payment Reconciliation worker 的独立 schema/工作流 | Gateway/Provider 提供事实；operator 审核；Payment 领域命令修复 | 第 17 章 | 发现与修复分权；冲突证据不能直接覆盖资金终态 |
| Settlement Line / Cost Attribution | Settlement 服务及其 schema | Router 只消费校准 artifact；Ledger/Finance 消费结果 | 第 18 章 | 原始 line append-only 且幂等；更正以新事实表达；不是 Payment 状态 |
| Region ownership/fencing epoch | Payment/全局最小协调存储 | API/worker 在每次副作用事务验证；Router 不拥有 Intent lease | 第 19 章 | 只有当前 epoch 可执行；DNS/进程存活不授予写权限 |
| Config Bundle | Routing Config Platform | Router 原子加载/ack；其他服务引用 digest | 第 20 章 | 完整 bundle 原子切换、last-known-good、作者审批分离、可回滚 |
| Model Artifact | Training/Model Registry 发布；Config Platform 审批；Router serving只读 | Replay/Experiment评估；Decision记录版本 | 第 21 章 | 不能越过 guardrail；unsupported/NaN/timeout 回退；需记录 propensity/support |
| Ledger | 外部 Ledger/Finance 系统 | Payment/Settlement 可发事实，Router 只引用 | 第 1、17、18 章边界 | Decision Store、Reconciliation 或 Settlement projection 都不能冒充会计账本 |

## 6. 批次发布顺序

“批次”是依赖波次，不是把多章压成一次大提交。批次内部仍按章节逐一执行：**代码仓增量与失败测试 → 全部代码门禁/CI → 固定实现 SHA → 完整正文、Mermaid/表格/必要 Ian 图 → delivery manifest → 书仓检查/推送**。

| 波次 | 章节 | 代码侧结束门禁 | 为什么必须按此顺序 |
| --- | --- | --- | --- |
| 0. 前置基线 | 第 1～3 章 | 当前 automatic-capture 纵向链路持续全绿；精确 Candidate/connection 从 Router 贯穿 Payment/Gateway/DB；Adapter 只执行不路由 | 后续所有 capability、key scope、health 和 attribution 都依赖可执行 connection，不可继续使用模糊 Provider string |
| 1. 决策输入与硬边界 | 第 4～6 章 | Context 无敏感凭证且严格版本化；Rule artifact 确定；Capability 三值过滤和两个同 PSP connection E2E | 先固定输入、Policy 语义和合法集合，才允许引入优化信号 |
| 2. 可解释优化 | 第 7～9 章 | 定点成本、稳健成功率、Risk guardrail 与稳定 score breakdown；Eligibility-over-score 失败测试 | Cost/Success/Risk 在同一 pipeline 收束，防止三章各自做选择器 |
| 3. 单笔可靠性 | 第 10～12 章 | UNKNOWN crash/restart、三层并发幂等、严格 no-side-effect fallback E2E；单路径回归持续通过 | 先证明原 operation 安全恢复，再允许创建第二 Attempt |
| 4. 跨交易可靠性与证据 | 第 13～16 章 | 局部 circuit/bulkhead、append-only Decision 投影、有限基数 telemetry、shadow 无副作用与稳定分桶 | Health、审计、观测建立后才有资格运行真钱策略实验 |
| 5. 事实与经济闭环 | 第 17～18 章 | checkpointed reconciliation、冲突 review、settlement line 幂等匹配和 Expected-vs-Actual | 先校正支付事实，再用其计算实际成本，避免反馈污染 |
| 6. 全球执行与治理 | 第 19～20 章 | 两 Region fencing/split-brain 测试；配置双人审批、原子 bundle、last-known-good/rollback | Multi-Region 明确本地/全局状态后，控制平面才能正确发布到各区域 |
| 7. 可选智能化 | 第 21 章 | OPE support 校验、shadow/challenger、model failure fallback、guardrail 不可绕过 | 只有 Decision/Experiment/Config/Actual outcome 都存在时，模型比较才有证据基础 |
| 8. 综合收束 | 第 22 章、附录 A～C、glossary 最终校正 | 四条系统故障旅程、race/vet/contracts/integration/Compose E2E、容量与 launch report；全部术语/状态/链接一致 | 最后只组合和审计已有机制，不在案例章或附录偷偷引入新核心能力 |

每一波次都保留上一波次回归；尤其是 merchant idempotency、UNKNOWN 单资金路径、冲突 `review_required`、production simulator fail-closed 和 PAN/CVV/SAD 禁止项。若本机没有真实 PostgreSQL/Compose 条件，不能省略测试，只能把最终门禁交给最新 GitHub Actions 并等待全绿。

## 7. 范围与重复自检

### 7.1 唯一主责检查

- Context 只在第 4 章首次设计；第 14 章只保存摘要和版本证据。
- Rule 求值只在第 5 章；Capability/Eligibility 只在第 6 章；Control Plane 生命周期只在第 20 章。
- Cost estimate 只在第 7 章；success estimate 只在第 8 章；Risk 与多目标选择只在第 9 章；第 21 章模型只能实现同一 scorer interface。
- Retry/UNKNOWN（第 10 章）、identity/idempotency（第 11 章）、new-Attempt Fallback（第 12 章）、future-request Health（第 13 章）是四种不同机制。
- Decision evidence（第 14 章）、measurement（第 15 章）、causal experiment（第 16 章）、fact correction（第 17 章）、economic attribution（第 18 章）不互相代替。
- Multi-Region ownership（第 19 章）与配置分发治理（第 20 章）分开；DNS、cache 或 config version 都不能替代 fencing。
- Provider Adapter 不新增章：第 3 章执行边界，第 6 章 capability publication。Risk-Aware 不新增章：全部收束在第 9 章。

### 7.2 不提前透支检查

- 第 4～8 章只产生 typed inputs/features，不提前实现 Decision Store、Experiment 或 ML。
- 第 10 章不创建新 Attempt；第 11 章不声称跨 PSP key 去重；第 12 章不把 UNKNOWN 当 fallback 条件。
- 第 13 章只处理技术健康，不用 Authorization Rate 触发 circuit；第 15 章的指标不自动改资金状态。
- 第 17 章自动修复只走领域状态机；第 18 章不创建 Ledger；第 20 章不允许在线实例拼接半套 artifact。
- 第 21 章没有可靠 propensity/support 时不得宣称因果提升；第 22 章的 100k TPS 是目标约束，只有实测证据才能成为容量结论。
- 附录只统一参考与门禁，不在目录末尾偷偷实现正文没有解释的新支付能力。

### 7.3 每章进入正文前的最小验收

1. 对照本文件写一页 chapter contract，并把“留给后章”原样转成 delivery manifest 的未实现范围。
2. 在代码中先补失败测试，再实现唯一纵向增量；Go 示例必须来自或可映射到真实包/API。
3. 检查本章术语在上表只有一个权威 owner，状态变化只由 owner 执行。
4. 在线核验 Provider、标准、法规和合规陈述，只使用对应官方/原始资料并把链接贴近主张。
5. 选择最小信息图：精确关系优先 Mermaid/表格；Ian 图只承担直觉和节奏，不能替代状态机、时序或数据表。
6. 更新 glossary、OpenAPI、README、architecture/ADR 和 delivery manifest；所有能力声明必须区分教学 Simulator、可运行支付核心、PSP sandbox 与 live system。
