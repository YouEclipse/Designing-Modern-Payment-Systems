# 第 7 章 Cost-Aware Routing

第 6 章完成 Eligibility 之后，Aurora 已经能排除不支持当前币种、Payment Method 或 operation 的通道。剩下的 Candidate 都“可以处理”这笔支付，但它们未必具有相同的经济成本。团队很自然地提出下一个问题：能否选择最便宜的一条？

问题出在“便宜”不是 Provider 网站上的一个百分比。固定处理费、比例费、跨境附加费、FX、卡组织与 Interchange、失败尝试计费方式，以及 Merchant Account 对应的合同版本，都可能改变比较结果。在线系统看到的也只是决策时可得信息；最终账单还要等 Capture、Clearing、Settlement 和 Provider 报告之后才能确认。

本章建立如下边界：

> **Expected Cost 是 Router 在决策时，使用版本化费率与汇率事实，对一个已经 eligible 的 Route Candidate 计算出的、可分解的成本估计。它只是一项决策特征，不是扣款、入账、Settlement 或利润的资金事实，也绝不覆盖事后归因得到的 Actual Cost。**

因此，本章不会用成本突破硬约束，不会因为另一条通道更便宜而改道执行 Capture、Void 或 Refund，也不会在 Provider 结果为 `UNKNOWN` 时创建一笔“更省钱”的新支付。第 8 章才引入成功概率，第 9 章才组合多目标评分，第 18 章再讨论 Settlement 与 Actual Cost Attribution。

配套 Go 系统已经把这一边界落成 Router v3 的窄纵向切片：`fee_catalog.v1` 按 `candidate_id + connection_revision + operation + payment_currency` 绑定不可变费率，支持 fixed/BPS、显式跨境分类、定向 FX Quote，以及 `attempt / success_only / failure_only` 三种互斥费用情景；每条 `cost.Result` 都随完整 Breakdown、Artifact version/digest 和求值时间进入 Optimization evidence。Scorer 当前明确使用 `SucceededCostMicros`，而不是把成功、失败费用相加，也不把它称为 Actual Cost。

这仍不是生产费率平台。当前开发配置来自仓库内的合成 fixture，没有实时 Provider 报价、合同同步、审批发布、阶梯价或 Settlement/Invoice 归因；它只能证明契约、算术、身份和降级语义。以下正文中的 Go 代码是为了逐步解释这些不变量而压缩过的教学投影；本章末尾会把每个简化字段映射回真实实现。

## 7.1 Problem：最低费率不一定是最低成本

假设两条 Candidate 都已通过 Eligibility，它们的合同费率如下。数字仅用于推导，不代表任何 Provider 的公开报价：

| Candidate | 精确执行身份 | 固定费 | 比例费 | 10.00 USD 的估计 | 100.00 USD 的估计 |
| --- | --- | ---: | ---: | ---: | ---: |
| `cand_us_a` | Provider A / `acct_us` / `conn_a` / `us-east` | 0.30 USD | 2.60% | 0.56 USD | 2.90 USD |
| `cand_us_b` | Provider B / `acct_us` / `conn_b` / `us-east` | 0.10 USD | 2.90% | 0.39 USD | 3.00 USD |

只比较百分比会始终选择 A，但小额支付上 B 更便宜；只比较固定费则会始终选择 B，但大额支付上 A 更便宜。价格交叉并不是边缘情况，而是“固定项 + 随金额变化项”的直接结果。Adyen 的官方价格页也把处理费与 Payment Method 费用分开，并说明部分卡交易的 Interchange++ 包含可变 Interchange 与其他组成项；这说明生产模型需要表达成本结构，而不是复制一个宣传页百分比。[Adyen Pricing](https://www.adyen.com/pricing)

即使把固定费加回来，`Provider A` 仍然不是正确的计价键。同一 Provider 下可能有两个 Merchant Account、不同 Region 或 Connector，分别绑定不同合同、结算币种和费率。Router 比较的是第 6 章产出的精确 Route Candidate，而不是 PSP 品牌。

一次在线估计可以先写成：`C_attempt = C_fixed + C_variable + C_crossborder + C_fx + C_other_known`。

这个式子描述“按当前信息发起一次 Attempt 的估计成本”，不等于“得到一笔成功支付的成本”。若一次便宜尝试更容易失败，随后还要重试或 Fallback，最终业务成本可能更高；但在第 8 章提供分段成功概率之前，本章不能凭空回答这个问题。

```mermaid
flowchart LR
    All["配置中的全部 Candidate"] --> Eligibility["第 6 章：Eligibility<br/>硬约束"]
    Eligibility -->|"eligible only"| Eligible["Eligible Candidate Set"]
    Eligibility -->|"ineligible / unknown"| Excluded["排除证据"]
    Eligible --> Resolve["按 Candidate + decision_at<br/>解析 Fee Schedule"]
    Resolve --> Estimate["Expected Cost<br/>逐项分解"]
    Estimate --> Order["成本偏好信号"]
    Order --> Later["第 8～9 章<br/>成功率与多目标组合"]
    Excluded -.->|"成本不能复活"| Order
```

图中最重要的箭头不是“Cost → Order”，而是“Eligibility → eligible only”。没有合同、违反区域要求或不支持当前 operation 的 Candidate，即使估计成本为零也不能进入比较。

![圆黑框眼镜小人拖着低费率门票后面不断展开的固定费、跨境和 FX 小票，把它们一起放上总成本秤](../assets/cost-aware-routing-illustrations/01-hidden-receipts.png)

## 7.2 Simple Design：从一个比例费开始

最小系统可以为每个精确 Candidate 配置一个 BPS（basis points，万分之一）比例费，只在已通过硬约束的集合中选择估计值最小者。下面的例子刻意保留简单模型的限制，但已经做到三件正确的事：金额使用最小货币单位整数；缺费率返回错误而不是当成零；稳定 Candidate ID 打破平局。

```go
package simplecost

import (
	"errors"
	"math"
	"sort"
)

var (
	ErrInvalidAmount = errors.New("invalid payment amount")
	ErrMissingRate   = errors.New("missing candidate rate")
	ErrOverflow      = errors.New("cost arithmetic overflow")
)

type Quote struct {
	CandidateID string
	FeeMinor    int64
}

// CheapestPercentage receives only candidates already accepted by Eligibility.
// Integer division currently rounds down; production rounding is added later.
func CheapestPercentage(
	amountMinor int64,
	eligibleCandidateIDs []string,
	rateBPS map[string]int64,
) (Quote, error) {
	if amountMinor <= 0 {
		return Quote{}, ErrInvalidAmount
	}
	quotes := make([]Quote, 0, len(eligibleCandidateIDs))
	for _, id := range eligibleCandidateIDs {
		bps, ok := rateBPS[id]
		if !ok || bps < 0 {
			return Quote{}, ErrMissingRate
		}
		if bps != 0 && amountMinor > math.MaxInt64/bps {
			return Quote{}, ErrOverflow
		}
		quotes = append(quotes, Quote{
			CandidateID: id,
			FeeMinor:    amountMinor * bps / 10_000,
		})
	}
	if len(quotes) == 0 {
		return Quote{}, ErrMissingRate
	}
	sort.Slice(quotes, func(i, j int) bool {
		if quotes[i].FeeMinor != quotes[j].FeeMinor {
			return quotes[i].FeeMinor < quotes[j].FeeMinor
		}
		return quotes[i].CandidateID < quotes[j].CandidateID
	})
	return quotes[0], nil
}
```

这个版本适用于费率完全同构、同币种且没有固定项的早期系统。它的价值不是精确，而是建立一个可替换的纯函数边界：输入是 eligible IDs、金额和 Router 自己持有的费率快照，输出是估计与稳定次序；函数不读 Payment 数据库，不调用 Gateway，也不改变 Payment State。

它也暴露了下一阶段的问题：整数除法究竟向下、向上还是四舍五入；费率何时生效；固定费以什么币种表示；跨境如何判定；FX 用哪一个时间点和来源；某条 Candidate 缺数据时是否还能比较。把这些问题留给“默认值”会得到一个看似稳定、实际不可审计的成本模型。

## 7.3 Problems at Scale：成本是带时间和来源的向量

### 7.3.1 每个成本项都有自己的可知性

生产费率不只是多几个字段。不同项目可能来自合同、卡组织规则、Provider 估算或事后账单，决策时的确定程度不同：

| 成本项 | 典型计价基础 | 决策时能否精确知道 | 在线模型的安全表达 |
| --- | --- | --- | --- |
| Processing fixed fee | 每次请求或交易的固定额 | 合同明确时通常可知 | 固定点金额 + 币种 + 适用 operation |
| Provider markup | 金额的 BPS、固定额或组合 | 合同和分类明确时可知 | BPS/固定额 + rounding mode |
| Interchange / Scheme fee | 卡产品、地区、MCC、清算条件等 | Authorization 时可能只能估算 | `estimated` component + basis/version |
| Cross-border surcharge | Provider/卡组织定义的地域组合 | 地理事实和规则都已知时可知 | 显式分类结果 + classification version |
| FX cost / markup | 原币、结算币、汇率与加价 | 取决于谁转换以及报价时效 | 有理数汇率 + as-of/expiry/source |
| Failed-attempt fee | 请求、Authorization 或其他事件 | 取决于合同与最终事件 | 只纳入有合同依据的场景，不猜测 |
| Refund / Chargeback fee | 后续独立事件 | 路由时通常尚未发生 | 作为单独情景或后续 Actual Cost，不伪装成当前扣款 |

卡费率为何难以压成单个常量，可以从卡组织的一手说明看出：Mastercard 的官方页面列出的适用因素包括 Merchant Category、Authorization 到 Clearing 的时间、提交的数据和交易量等，并提醒具体协议问题应与 Acquirer 或支付服务方确认。[Mastercard Merchant Interchange Rates](https://www.mastercard.com/us/en/business/support/merchant-interchange-rates.html) 因此，公开费率表可以帮助理解结构，却不能取代 Aurora 与某个 Merchant Account 的合同 Artifact。

`CostBreakdown` 必须保留每一项的名称、金额、币种、估算依据和版本。总数只是这些项在同一 quote currency 下的确定性求和；若无法把某项合法换算到 quote currency，就没有可比较的总数，而不是“总数暂时少一点”。

### 7.3.2 费率必须绑定 Candidate、合同版本和有效时间

费率发布是时间区间，不是覆盖数据库中的一行。假设旧合同在 `2026-09-01T00:00:00Z` 结束，新合同同时开始，系统应采用半开区间 `[effective_from, effective_until)`：

| `decision_at` | 命中的 Schedule | 原因 |
| --- | --- | --- |
| `2026-08-31T23:59:59.999Z` | `fee_v17` | 仍在旧区间内 |
| `2026-09-01T00:00:00Z` | `fee_v18` | 旧区间右端不包含，新区间左端包含 |
| 两个版本都命中 | 配置无效 | 不能靠加载顺序选择 |
| 没有版本命中 | `missing_schedule` | 不能沿用“最近一个”而不留证据 |

一次决定只能读取一次 `decision_at`，并记录 Schedule version 与内容 digest。若对 A 估价后调用 `time.Now()`，恰逢版本切换，再对 B 使用新时间，就会比较两个从未同时有效的合同。已发布 Artifact 应不可变；修正错误要发布新版本，历史 Expected Cost 仍指向当时使用的旧 Artifact。

阶梯价和月度承诺量更棘手。“本月第十万笔交易”可能在多个 Region 并发到达，在线计数也会因撤销、补录和账单口径变化而修正。本章的安全做法是让 Control Plane 发布一个有来源的当前 tier snapshot，或采用经批准的保守档位；它仍是 Expected Cost。逐张 Invoice、阶梯校准和实际分摊留到第 18 章，发布审批与配置治理留到第 20 章。

### 7.3.3 币种代码、精度和舍入都属于合同

`amount_minor = 1000` 只有与币种及 minor-unit exponent 一起才有意义。ISO 4217 为币种提供三字母代码，并给出有 minor unit 的币种与主单位之间的关系；参考数据会维护更新，因此 Router 应消费版本化 currency metadata，而不是假设所有币种都有两位小数。[ISO 4217 Currency Codes](https://www.iso.org/iso-4217-currency-codes.html)

为了表达低于一个最小货币单位的费用，本章使用 `cost_micros`：quote currency 主单位的百万分之一。例如 quote currency 为 USD 时，`1_000_000 cost_micros = 1 USD`。这是平台内部固定点单位，不是 Provider API 的通用格式，也不是“微分币”。所有边界都必须携带 `quote_currency`，UI 或报表只能在最后一步按声明的 rounding mode 展示。

若费用币种与 quote currency 不同，FX Quote 至少要包含 base/quote、整数 numerator/denominator、source、`as_of`、`expires_at` 和 version。浮点 `0.1` 不能进入费用路径。还要区分参考基准与真实成交汇率：欧洲央行明确说明其 Euro reference rates 仅供信息使用，并强烈不建议用于交易目的，所以它可以是对照或校准来源，却不能未经合同授权就冒充 Provider 实际 FX 价格。[ECB Euro Foreign Exchange Reference Rates](https://www.ecb.europa.eu/stats/policy_and_exchange_rates/euro_reference_exchange_rates/html/index.en.html)

### 7.3.4 最低单次成本不等于最低成功支付成本

Candidate A 的一次 Authorization 估计 20 cost cents，Candidate B 为 24 cost cents，本章只能说 A 的单次已知成本更低。若 A 的分段成功率更低、技术错误会产生额外尝试，或者某些失败也计费，支付最终完成所需的期望尝试链就不同。

这里不能偷偷加入一个全局成功率，也不应直接输出 `cost / probability`：概率的分桶、窗口、样本量与置信度尚未建立，Fallback 还受结果语义和 `UNKNOWN` 约束。第 7 章输出纯成本分解；第 8 章独立输出成功率信号；第 9 章再以显式目标和 Guardrail 组合两者。这样，成本配置错误不会伪装成资金失败，成功率模型漂移也不会改写历史费率事实。

### 7.3.5 Expected 与 Actual 必须并排存在

| 维度 | Expected Cost | Actual Cost |
| --- | --- | --- |
| 产生时间 | Routing Decision 时 | Provider 报告、Invoice、Clearing/Settlement 数据到达后 |
| 依据 | Fee/FX/Classification Artifact 快照 | 实际费用行、结算行和合同归因规则 |
| 主要用途 | 在 eligible Candidate 间提供决策信号 | 对账、财务归因、校准与模型误差分析 |
| 是否可回写历史值 | 不可；保留原 estimate/version | 可追加修正版本，但不覆盖原 Expected |
| 是否证明资金已结算 | 否 | 单一费用行也不当然证明整个 Payment 已结算 |

Provider 的交易级报表会出现事前难以完全确定的差异。以 Adyen 官方 Settlement details report 为例，文档说明报表可用于查看交易级成本，并列出 Commission、Markup、Scheme Fees、Interchange、Exchange Rate 等字段；这些是 Actual Cost Attribution 的输入证据，而不是 Router 在 Authorization 前已经知道的事实。[Adyen Settlement Details Report](https://docs.adyen.com/reporting/settlement-reconciliation/transaction-level/settlement-details-report)

校准流程可以比较 `actual - expected`，定位漏项、错误分类或过期费率，但修正方式是发布新 Artifact 并保留差异证据，不是把历史 Expected Cost 改成 Actual。否则团队将永远无法知道当时为何选择那条 Route。

## 7.4 Production Design：可解释、定点且可拒绝比较的估价器

### 7.4.1 数据流先过滤，再估价

生产 Cost Evaluator 消费两个冻结输入：第 4 章的 Routing Context 中与金额、币种、地域分类相关的最小事实，以及第 6 章给出的 eligible Candidate Set。它还读取 Router-owned 的 Fee/FX Artifact 快照，但不向 Provider 临时询价，不读取 Payment Method Token，也不接触 PAN 或 CVV。

```mermaid
sequenceDiagram
    autonumber
    participant O as Payment Orchestrator
    participant R as Router
    participant E as Eligibility
    participant C as Cost Evaluator
    participant F as Fee/FX Snapshot

    O->>R: Decide(context, decision_at)
    R->>E: Filter exact candidates
    E-->>R: eligible set + exclusion reasons
    R->>F: Resolve immutable artifact snapshot
    F-->>R: versions + digests
    R->>C: Estimate eligible set only
    C-->>R: breakdowns or explicit unavailable status
    R-->>O: candidate plan + cost evidence
    Note over O,R: No Provider execution and no Payment State transition
```

如果一条 Candidate 在 Eligibility 阶段被排除，它连 `FeeSchedule` 都不应参与排序。反过来，Cost Evaluator 缺数据也不会暗中改变能力事实：它返回 `missing_schedule`、`stale_fx` 或 `unknown_cross_border`。只有单独、明确发布的 Guardrail 才能规定“成本不可比较时拒绝路由”；普通成本代码不能用 `+∞` 把 eligible Candidate 伪装成 ineligible。

### 7.4.2 Versioned FeeSchedule 与逐项 CostBreakdown

下面的 Go 代码实现一个同币种估价核心：固定费、BPS 比例费和显式跨境附加费都用 `cost_micros`；有效期采用半开区间；乘除通过商与余数拆分，避免先做可能溢出的 `amount * bps`。示例仅接受已规范化的 currency exponent，跨币种时返回 `ErrFXRequired`，由携带完整 provenance 的 FX 层处理。

这是教学简化，不是把仓库实现复制一遍：示例的 `FeeSchedule` 省略了真实键中的 `ConnectionRevision`、`Operation`、`PaymentCurrency`、`FixedCurrency`，也把真实的 `Attempt`、`Success`、`Failure` 三组 `Fee` 压成一个总额。生产 `cost.Estimate` 分别返回 `AttemptCostMicros`、`SucceededCostMicros` 和 `FailedCostMicros`；当前 v3 scorer 只消费成功情景的 `SucceededCostMicros`，原始三情景与 Breakdown 仍完整保留。

```go
package cost

import (
	"errors"
	"math"
	"time"
)

var (
	ErrScheduleNotEffective = errors.New("fee schedule is not effective")
	ErrUnknownCrossBorder   = errors.New("cross-border classification is unknown")
	ErrFXRequired           = errors.New("currency conversion is required")
	ErrInvalidCostInput     = errors.New("invalid cost input")
	ErrCostOverflow         = errors.New("cost arithmetic overflow")
)

type RoundingMode string

const (
	RoundDown   RoundingMode = "down"
	RoundHalfUp RoundingMode = "half_up"
	RoundUp     RoundingMode = "up"
)

type FactState string

const (
	FactKnown   FactState = "known"
	FactUnknown FactState = "unknown"
)

type CrossBorderFact struct {
	State       FactState
	Value       bool
	RuleVersion string
}

type FeeSchedule struct {
	ScheduleID      string
	ArtifactVersion string
	CandidateID     string
	QuoteCurrency   string
	EffectiveFrom   time.Time
	EffectiveUntil  time.Time
	FixedMicros     int64
	VariableBPS     int64
	CrossBorderBPS  int64
	PercentageRound RoundingMode
}

type Input struct {
	AmountMinor      int64
	PaymentCurrency  string
	CurrencyExponent uint8
	DecisionAt       time.Time
	CrossBorder      CrossBorderFact
}

type BreakdownItem struct {
	Kind         string
	AmountMicros int64
	Basis        string
}

type Estimate struct {
	CandidateID        string
	QuoteCurrency      string
	TotalMicros        int64
	Breakdown          []BreakdownItem
	ScheduleID         string
	ArtifactVersion    string
	CrossBorderRuleVer string
	EvaluatedAt        time.Time
}

func EstimateSameCurrency(schedule FeeSchedule, in Input) (Estimate, error) {
	if in.AmountMinor <= 0 || schedule.FixedMicros < 0 ||
		schedule.VariableBPS < 0 || schedule.CrossBorderBPS < 0 {
		return Estimate{}, ErrInvalidCostInput
	}
	if in.DecisionAt.Before(schedule.EffectiveFrom) ||
		(!schedule.EffectiveUntil.IsZero() && !in.DecisionAt.Before(schedule.EffectiveUntil)) {
		return Estimate{}, ErrScheduleNotEffective
	}
	if in.PaymentCurrency != schedule.QuoteCurrency {
		return Estimate{}, ErrFXRequired
	}
	if in.CurrencyExponent > 6 {
		return Estimate{}, ErrInvalidCostInput
	}

	factor, err := pow10(6 - in.CurrencyExponent)
	if err != nil || in.AmountMinor > math.MaxInt64/factor {
		return Estimate{}, ErrCostOverflow
	}
	amountMicros := in.AmountMinor * factor
	variable, err := mulDiv(amountMicros, schedule.VariableBPS, 10_000, schedule.PercentageRound)
	if err != nil {
		return Estimate{}, err
	}

	items := []BreakdownItem{
		{Kind: "fixed", AmountMicros: schedule.FixedMicros, Basis: "contract"},
		{Kind: "variable", AmountMicros: variable, Basis: "amount_bps"},
	}
	total, err := checkedAdd(schedule.FixedMicros, variable)
	if err != nil {
		return Estimate{}, err
	}

	if schedule.CrossBorderBPS > 0 {
		if in.CrossBorder.State != FactKnown {
			return Estimate{}, ErrUnknownCrossBorder
		}
		if in.CrossBorder.Value {
			surcharge, calcErr := mulDiv(
				amountMicros,
				schedule.CrossBorderBPS,
				10_000,
				schedule.PercentageRound,
			)
			if calcErr != nil {
				return Estimate{}, calcErr
			}
			total, err = checkedAdd(total, surcharge)
			if err != nil {
				return Estimate{}, err
			}
			items = append(items, BreakdownItem{
				Kind:         "cross_border",
				AmountMicros: surcharge,
				Basis:        "classified_by:" + in.CrossBorder.RuleVersion,
			})
		}
	}

	return Estimate{
		CandidateID:        schedule.CandidateID,
		QuoteCurrency:      schedule.QuoteCurrency,
		TotalMicros:        total,
		Breakdown:          items,
		ScheduleID:         schedule.ScheduleID,
		ArtifactVersion:    schedule.ArtifactVersion,
		CrossBorderRuleVer: in.CrossBorder.RuleVersion,
		EvaluatedAt:        in.DecisionAt,
	}, nil
}

func mulDiv(value, numerator, denominator int64, mode RoundingMode) (int64, error) {
	if value < 0 || numerator < 0 || denominator <= 0 {
		return 0, ErrInvalidCostInput
	}
	q, r := value/denominator, value%denominator
	if numerator != 0 && (q > math.MaxInt64/numerator || r > math.MaxInt64/numerator) {
		return 0, ErrCostOverflow
	}
	whole := q * numerator
	tail := r * numerator
	partial, remainder := tail/denominator, tail%denominator
	if mode == RoundUp && remainder > 0 {
		partial++
	} else if mode == RoundHalfUp && remainder >= denominator/2+denominator%2 {
		partial++
	} else if mode != RoundDown && mode != RoundHalfUp && mode != RoundUp {
		return 0, ErrInvalidCostInput
	}
	return checkedAdd(whole, partial)
}

func checkedAdd(left, right int64) (int64, error) {
	if left < 0 || right < 0 || right > math.MaxInt64-left {
		return 0, ErrCostOverflow
	}
	return left + right, nil
}

func pow10(exponent uint8) (int64, error) {
	value := int64(1)
	for i := uint8(0); i < exponent; i++ {
		if value > math.MaxInt64/10 {
			return 0, ErrCostOverflow
		}
		value *= 10
	}
	return value, nil
}
```

这段实现故意不在估价器里推断 `CrossBorder`。跨境含义可能依赖 Merchant Entity、Acquirer Region、Issuer Country 和合同定义；一个来自第 4 章 Context 与版本化分类规则的结果，必须连同 `RuleVersion` 一起进入。`customer_country != merchant_country` 不是通用真理。

`FeeSchedule` 的发布校验至少应证明：Candidate ID 存在且精确绑定 Merchant Account/Connector/Region；相同作用域的有效区间不重叠；币种与 exponent 来自允许的参考数据版本；所有适用组件声明 rounding；在平台允许的最大金额上不会溢出；version 与 canonical digest 一一对应。运行时验证不是配置发布校验的替代品，两道门都要保留。

### 7.4.3 Missing 和 Stale 不能变成免费

估价结果需要 `comparable`、`unavailable` 和原因，而不是只返回一个整数。尤其不能让 Go map 的零值把缺失配置变成最优路线：

| 情况 | 禁止行为 | 显式结果 | 可配置的安全动作 |
| --- | --- | --- | --- |
| Candidate 没有有效 Schedule | `cost = 0` | `missing_schedule` | 整个成本阶段 abstain，或按已批准静态策略处理 |
| Schedule 已过期 | 沿用“最后一版” | `stale_schedule` | fail closed 或退出成本比较 |
| 需要 FX 但 Quote 缺失/过期 | 用 1:1、NaN 或缓存零值 | `missing_fx` / `stale_fx` | 不生成 total；等待或降级 |
| Cross-border 适用但事实 unknown | 默认 domestic | `unknown_cross_border` | 不比较该估值，记录数据缺口 |
| 算术溢出或非法舍入 | wrap 成负数 | `arithmetic_error` | 拒绝该次成本求值并告警 |
| 只有部分 Candidate 可估价 | 只比较有数据者且不留痕 | `incomplete_comparison` | 按 Merchant policy 对整个阶段统一处置 |

“退出成本比较”不等于随机选择。它应回到明确、版本化且仍受 Eligibility 限制的静态 Candidate order，并记录 `COST_STAGE_ABSTAINED`；若 Merchant Policy 要求所有候选成本可比，Router 则返回受控的无决定结果。哪种策略更合适取决于可用性目标，但必须由配置声明，不能由异常分支临时决定。

如果业务确实要求“缺少可信费率的 Candidate 禁止执行”，应把 `cost_data_required` 做成独立 Guardrail，并给出稳定排除原因。这样审计者能区分“该通道没有能力处理”和“组织不接受无法计价的通道”，而不是看到 Cost Evaluator 悄悄把未知值改成无穷大。

### 7.4.4 只对完整估值排序

排序函数必须接收与 eligible set 一一对应的完整估值。它不能自己读取配置，也不能忽略缺项：

```go
package costrank

import (
	"errors"
	"sort"
)

var ErrIncompleteComparison = errors.New("incomplete cost comparison")

type Estimate struct {
	CandidateID   string
	TotalMicros   int64
	QuoteCurrency string
}

func RankEligible(eligibleIDs []string, estimates map[string]Estimate) ([]Estimate, error) {
	if len(eligibleIDs) == 0 {
		return nil, ErrIncompleteComparison
	}
	ranked := make([]Estimate, 0, len(eligibleIDs))
	seen := make(map[string]struct{}, len(eligibleIDs))
	quoteCurrency := ""
	for _, id := range eligibleIDs {
		if _, duplicate := seen[id]; id == "" || duplicate {
			return nil, ErrIncompleteComparison
		}
		seen[id] = struct{}{}
		estimate, ok := estimates[id]
		if !ok || estimate.CandidateID != id || estimate.TotalMicros < 0 {
			return nil, ErrIncompleteComparison
		}
		if quoteCurrency == "" {
			quoteCurrency = estimate.QuoteCurrency
		}
		if estimate.QuoteCurrency == "" || estimate.QuoteCurrency != quoteCurrency {
			return nil, ErrIncompleteComparison
		}
		ranked = append(ranked, estimate)
	}
	if len(ranked) != len(estimates) {
		return nil, ErrIncompleteComparison
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].TotalMicros != ranked[j].TotalMicros {
			return ranked[i].TotalMicros < ranked[j].TotalMicros
		}
		return ranked[i].CandidateID < ranked[j].CandidateID
	})
	return ranked, nil
}
```

Candidate ID 作为最终 tie-breaker 只保证确定性，不表达业务偏好。若成本差小于合同精度或运营设定的 materiality threshold，可由第 9 章的评分策略决定是否视为平局；本章不把成功率、风险或延迟偷偷塞进成本数字。

### 7.4.5 决策证据与资金证据走不同链路

```mermaid
flowchart TB
    Decision["Routing Decision<br/>Expected Cost + artifact versions"] --> Execute["Gateway 执行原始 Provider Operation"]
    Execute --> Outcome{"Provider Outcome"}
    Outcome -->|"明确成功/失败"| PaymentState["Orchestrator 更新 Payment/Attempt"]
    Outcome -->|"timeout / ambiguous"| Unknown["UNKNOWN / possible<br/>查询原 operation"]
    Unknown -->|"证据冲突，或 deadline / budget 耗尽"| Review["integrity_status=review_required"]
    PaymentState --> LaterOps["Capture / Void / Refund<br/>固定原 provider connection + reference"]

    Reports["Provider reports / invoices / settlement lines"] --> Attribution["第 18 章 Actual Cost Attribution"]
    Attribution --> Compare["actual - expected<br/>校准证据"]
    Compare --> NewArtifact["发布新的 Fee Artifact"]
    NewArtifact -.->|"不覆盖历史 estimate"| Decision
```

Expected Cost 不会让 Payment 变成 `AUTHORIZED` 或 `CAPTURED`，Actual Cost 也不能单凭一条费用行把 Payment 标成 `SETTLED`。若原始调用超时，成本更低的另一个 Candidate 不是解决不确定性的工具；系统仍要用原 `operation_id` 查询、接收 webhook 或对账。普通 `UNKNOWN/possible` 先保持 `processing` 并继续原操作恢复；只有资金证据冲突，或自动恢复的 deadline / budget 耗尽而仍无法安全解释时，才把独立的完整性状态置为 `review_required`。不能把每个暂时未知都预先送入人工复核。

同理，后续 Capture、Void 和 Refund 是原支付链上的操作，必须使用原 Provider connection 与 Provider reference。重新跑一次 Cost-Aware Routing 也许会选出更便宜的 Provider，却无法把另一家 Provider 持有的 Authorization 搬过去；这么做只会创建新的资金事实并破坏历史。

## 7.5 Failure Modes：成本优化如何把系统带偏

| 失败模式 | 表面现象 | 根因与控制 |
| --- | --- | --- |
| 先算成本、后做 Eligibility | 不支持币种的 Candidate 因便宜胜出 | 先冻结 eligible set，Cost 只消费该集合 |
| 费率只按 Provider 品牌配置 | 同一 PSP 的不同账户被算成同价 | 以精确 Candidate/合同作用域为键 |
| 金额或汇率使用 `float64` | 边界金额出现跨语言差异 | minor unit + fixed point + 显式 rounding |
| Missing/NaN/解析失败变成零 | 数据最差的 Candidate 反而最便宜 | 显式 unavailable，永不默认免费 |
| 两个费率版本同时有效 | 重启或节点顺序改变结果 | 发布时拒绝重叠区间，记录 digest |
| 把旧 FX Quote 无限复用 | 波动时成本严重偏离 | `as_of`、expiry、source 与降级策略 |
| 用一个 `country` 判断跨境 | 国内交易被加费或跨境漏费 | 命名地理事实，版本化分类规则 |
| 把所有失败都假设为免费或收费 | 估值系统性偏差 | 只按合同与事件口径纳入，保留 basis |
| 用 Actual 覆盖 Expected | 历史决定看起来永远“正确” | 两条记录不可变关联，以差值校准 |
| 为省钱跨 Provider Capture/Refund | 原 Authorization 与后续操作断链 | 后续资金操作固定原 connection/reference |
| `UNKNOWN` 后立即选更便宜通道 | 顾客可能被重复扣款 | 原 operation reconciliation，禁止成本驱动 Fallback |
| 把成本分数当利润或账本 | Router 数字被用于财务入账 | Expected 只作决策证据，Ledger/Settlement 各有权威来源 |

还要区分两种“未知”：`cost_estimate_unavailable` 表示无法可靠比较费用；Payment Attempt 的 `UNKNOWN` 表示 Provider 操作结果不确定。前者可以让成本阶段 abstain，后者则触发原操作查询与对账。用同一个布尔字段表达两者，会让配置缺失升级成资金状态事故。

测试应围绕可证伪的不变量，而不只验证一个 happy path 金额：低比例高固定费在小额交易上应输给高比例低固定费；恰好位于 `effective_from` 的请求必须只命中新版本；平台最大允许金额不得溢出；缺 FX、缺 Schedule 或 unknown cross-border 绝不能产生零、负数或 NaN；把一个 ineligible Candidate 配成极低费率也不能让它回到结果集。

## 7.6 Trade-offs：精确、可用与复杂度

| 选择 | 收益 | 代价 | 适用边界 |
| --- | --- | --- | --- |
| 每个 Candidate 都要求完整估值 | 比较公平，最易审计 | 单个数据缺口可能让成本阶段不可用 | 合同数据成熟、强可比场景 |
| 缺数据时回退静态顺序 | 保留支付可用性 | 暂时放弃成本收益 | 必须记录 abstain，仍经过 Eligibility |
| 统一 quote currency | 排序直接、解释一致 | 依赖 FX Quote 与额外舍入 | 多结算币种平台 |
| 各币种分别排序 | 减少不必要换算 | 不同币种无法形成总序 | 候选天然同币种分组时 |
| 完整复制复杂费率表 | 理论拟合更高 | 发布、测试、解释和漂移成本高 | 只有 Actual 数据能持续校准时 |
| 保守 tier snapshot | 在线确定、易重放 | 可能错过短期折扣 | 阶梯计费但并发口径不稳定时 |
| 逐请求实时询价 Provider | 可能接近最新报价 | 增加延迟、外部依赖与不可重放性 | 通常不放在同步 Router 热路径 |

负成本或 rebate 也需要明确选择。有些合同可能在月末产生返利，但把预计返利直接写成负的单笔费用，会让 Router 为追逐尚未确认的奖励集中流量。更安全的起点是在线组件非负，返利在第 18 章按可审计规则归因；若未来确需纳入，则要声明封顶、适用区间、置信度和总成本下限，不能只解除 `>= 0` 校验。

成本模型越复杂，越容易给出一个精确到六位小数、却建立在错误分类上的数字。解释性不是附加 UI：Breakdown、basis、Schedule/FX/classification version 和 unavailable reason 都是模型本身。只存 `total = 18342`，就无法判断差异来自新费率、跨境判定、FX 过期还是舍入规则。

最后，最低 Expected Cost 不是最终业务目标。成功率、风险、Provider Health、延迟与流量承诺可能让系统选择并非最便宜的 Candidate；这不是成本模型失败，而是职责正确。Cost Evaluator 应忠实回答“在当前假设下这次尝试估计花多少”，把“应该选谁”留给显式的多目标决策层。

## 7.7 配套 Go 系统：已实现的 v3 成本边界

本章的[固定实现](https://github.com/YouEclipse/payments/tree/bc4afde24da708d22f67da72a90a1557b1477173/services/router/internal/cost)已经提供 `Compile(Artifact)` 与 `Estimate(at, Query)`。编译阶段拒绝 Candidate revision 漂移、Schedule 重叠、未知币种精度、重叠 FX 区间和非法舍入；求值阶段只接受显式冻结的时间、已允许的精确连接、命名过的跨境分类与 quote currency。Missing、Stale、Unknown 和 Overflow 都返回 typed state/reason，失败时 `Estimate=nil`，合法零费用仍保留合同证据。纯 Cost package 可表达 `sale` 与 `authorize` Schedule，但当前 Router v3 Optimization 只承认 `sale + automatic` Cohort；这不代表 manual Capture 资金链已经实现。

Router 的 `routing_optimization.v1` 要求 Fee、Performance 与 Capability 三份 Candidate manifest 在 ID 和 revision 上完全一致。它只为 Eligibility、Rule 和 Risk 已允许的集合估价；任一 Candidate 的费用或表现不可用时，Scoring Profile 只能对整批执行已经声明的 `static_order` 或 `reject`，不能只比较数据完整者。最终 `executable_until` 还会被所用 Schedule、FX、跨境分类、Performance window、Risk 与 Scoring Profile 的最早失效时间缩短，避免一份曾经可算的价格无限授权未来执行。

当前开发 Artifact 是合成数据，Provider 也是 Simulator；代码测试证明的是确定性、可解释性和失败边界，不证明真实 PSP 合同价格、实时 FX 或账单精度。真实费率摄取与审批发布留给第 20 章，Actual Cost 与 Settlement 归因仍留给第 18 章。

## 7.8 Summary：把成本变成证据，而不是资金真相

本章从一个比例费配置出发，先遇到固定费导致的价格交叉，再加入跨境、FX、费率有效期、失败尝试和合同作用域，最终得到一组生产不变量：

1. Eligibility 和 Guardrail 永远先于成本；成本不能复活被排除的 Candidate。
2. 费率绑定精确 Candidate、Merchant Account/Connector/Region 作用域、有效时间和不可变版本，而不是 Provider 品牌。
3. 金额、BPS、FX 和舍入使用声明过单位的整数/固定点运算；任何路径都不用浮点。
4. `CostEstimate` 必须包含 quote currency、逐项 Breakdown、basis、Schedule/FX/classification version 和求值时间。
5. Missing、Stale、Unknown 与 Overflow 都是显式不可比较状态，绝不变成零费用、NaN、负数或静默默认值。
6. Expected Cost 是决策时快照；Actual Cost 来自后续报告与归因。二者并排保存，用差异校准，不互相覆盖。
7. 最低单次成本不等于最低成功支付成本。本章不借用未知成功率，第 8 章和第 9 章再组合目标。
8. 成本从不改变 Auth、Capture、Settlement 或 `UNKNOWN` 语义；Capture、Void、Refund 与不确定结果处理始终固定在原 Provider operation 链上。

做到这些，Router 才不是“挑一个网页价格最低的 PSP”，而是在合法候选之间产生可解释、可重放、能被 Actual 数据持续校准的经济信号。下一章将在不污染这条成本证据链的前提下，回答另一个独立问题：当前细分交易在每条 Candidate 上成功的概率有多大。

## 延伸阅读

* [第 4 章：构建 Routing Context](routing-context.md)
* [第 6 章：Provider Eligibility](provider-eligibility.md)
* [第 8 章：Success-Rate Routing](success-rate-routing.md)
* [第 9 章：Provider Scoring 与 Risk-Aware Routing](provider-scoring.md)
* [第 18 章：Settlement 与 Cost Attribution](../part-4-operating-at-scale/settlement.md)
* [ISO 4217 Currency Codes](https://www.iso.org/iso-4217-currency-codes.html)
* [ECB Euro Foreign Exchange Reference Rates](https://www.ecb.europa.eu/stats/policy_and_exchange_rates/euro_reference_exchange_rates/html/index.en.html)
* [Adyen Settlement Details Report](https://docs.adyen.com/reporting/settlement-reconciliation/transaction-level/settlement-details-report)
